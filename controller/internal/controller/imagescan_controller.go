package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	invulnerablev1alpha1 "github.com/pacokleitz/invulnerable/controller/api/v1alpha1"
)

const (
	imageScanFinalizer = "invulnerable.io/finalizer"
	conditionTypeReady = "Ready"
)

// ImageScanReconciler reconciles an ImageScan object
type ImageScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=invulnerable.io,resources=imagescans,verbs=get;list;watch
// +kubebuilder:rbac:groups=invulnerable.io,resources=imagescans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=invulnerable.io,resources=imagescans/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Note: Controller does NOT have create/delete permissions for ImageScans.
// Users create ImageScans, controller reconciles them.
// Controller CAN create/delete CronJobs (owned resources).
// For namespace-scoped deployment, use Role instead of ClusterRole.

// Reconcile is part of the main kubernetes reconciliation loop
func (r *ImageScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the ImageScan instance
	imageScan := &invulnerablev1alpha1.ImageScan{}
	if err := r.Get(ctx, req.NamespacedName, imageScan); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request
			logger.Info("ImageScan resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ImageScan")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !imageScan.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, imageScan)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(imageScan, imageScanFinalizer) {
		controllerutil.AddFinalizer(imageScan, imageScanFinalizer)
		if err := r.Update(ctx, imageScan); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Get or create CronJob
	cronJob, err := r.reconcileCronJob(ctx, imageScan)
	if err != nil {
		logger.Error(err, "Failed to reconcile CronJob")
		r.setCondition(imageScan, conditionTypeReady, metav1.ConditionFalse, "ReconcileFailed", err.Error())
		if statusErr := r.Status().Update(ctx, imageScan); statusErr != nil {
			logger.Error(statusErr, "Failed to update ImageScan status")
		}
		return ctrl.Result{}, err
	}

	// Update status
	imageScan.Status.CronJobName = cronJob.Name
	imageScan.Status.ObservedGeneration = imageScan.Generation
	r.setCondition(imageScan, conditionTypeReady, metav1.ConditionTrue, "ReconcileSuccess", "CronJob successfully reconciled")

	if err := r.Status().Update(ctx, imageScan); err != nil {
		logger.Error(err, "Failed to update ImageScan status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully reconciled ImageScan", "cronJob", cronJob.Name)
	return ctrl.Result{}, nil
}

// reconcileCronJob creates or updates the CronJob for an ImageScan
func (r *ImageScanReconciler) reconcileCronJob(ctx context.Context, imageScan *invulnerablev1alpha1.ImageScan) (*batchv1.CronJob, error) {
	logger := log.FromContext(ctx)

	cronJobName := fmt.Sprintf("%s-scanner", imageScan.Name)
	cronJob := &batchv1.CronJob{}
	err := r.Get(ctx, types.NamespacedName{Name: cronJobName, Namespace: imageScan.Namespace}, cronJob)

	// Set defaults
	sbomFormat := imageScan.Spec.SBOMFormat
	if sbomFormat == "" {
		sbomFormat = "cyclonedx"
	}

	successfulJobsHistoryLimit := int32(3)
	if imageScan.Spec.SuccessfulJobsHistoryLimit != nil {
		successfulJobsHistoryLimit = *imageScan.Spec.SuccessfulJobsHistoryLimit
	}

	failedJobsHistoryLimit := int32(3)
	if imageScan.Spec.FailedJobsHistoryLimit != nil {
		failedJobsHistoryLimit = *imageScan.Spec.FailedJobsHistoryLimit
	}

	scannerImage := "invulnerable-scanner:latest"
	pullPolicy := corev1.PullIfNotPresent
	if imageScan.Spec.ScannerImage != nil {
		repo := imageScan.Spec.ScannerImage.Repository
		if repo == "" {
			repo = "invulnerable-scanner"
		}
		tag := imageScan.Spec.ScannerImage.Tag
		if tag == "" {
			tag = "latest"
		}
		scannerImage = fmt.Sprintf("%s:%s", repo, tag)
		if imageScan.Spec.ScannerImage.PullPolicy != "" {
			pullPolicy = imageScan.Spec.ScannerImage.PullPolicy
		}
	}

	apiEndpoint := imageScan.Spec.APIEndpoint
	if apiEndpoint == "" {
		// Default to backend service in same namespace
		apiEndpoint = fmt.Sprintf("http://invulnerable-backend.%s.svc.cluster.local:8080", imageScan.Namespace)
	}

	workspaceSize := imageScan.Spec.WorkspaceSize
	if workspaceSize == "" {
		workspaceSize = "10Gi"
	}

	// Define the desired CronJob
	desiredCronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJobName,
			Namespace: imageScan.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "invulnerable-scanner",
				"app.kubernetes.io/instance":   imageScan.Name,
				"app.kubernetes.io/component":  "scanner",
				"app.kubernetes.io/managed-by": "invulnerable-controller",
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:                   imageScan.Spec.Schedule,
			TimeZone:                   imageScan.Spec.TimeZone,
			Suspend:                    &imageScan.Spec.Suspend,
			SuccessfulJobsHistoryLimit: &successfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     &failedJobsHistoryLimit,
			ConcurrencyPolicy:          batchv1.ForbidConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name":      "invulnerable-scanner",
								"app.kubernetes.io/instance":  imageScan.Name,
								"app.kubernetes.io/component": "scanner",
							},
						},
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyOnFailure,
							SecurityContext: &corev1.PodSecurityContext{
								RunAsNonRoot: ptr(true),
								RunAsUser:    ptr(int64(1000)),
								RunAsGroup:   ptr(int64(1000)),
								FSGroup:      ptr(int64(1000)),
							},
							Containers: []corev1.Container{
								{
									Name:            "scanner",
									Image:           scannerImage,
									ImagePullPolicy: pullPolicy,
									Env:             buildEnvVars(imageScan, apiEndpoint, sbomFormat),
									SecurityContext: &corev1.SecurityContext{
										AllowPrivilegeEscalation: ptr(false),
										Capabilities: &corev1.Capabilities{
											Drop: []corev1.Capability{"ALL"},
										},
										ReadOnlyRootFilesystem: ptr(false),
										RunAsNonRoot:           ptr(true),
										RunAsUser:              ptr(int64(1000)),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "scan-workspace",
											MountPath: "/tmp/syft",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "scan-workspace",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{
											SizeLimit: resourceQuantityPtr(workspaceSize),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Set resources if specified
	if imageScan.Spec.Resources != nil {
		desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources = *imageScan.Spec.Resources
	}

	// Set ImagePullSecrets if specified (allows pulling private scanner image)
	if len(imageScan.Spec.ImagePullSecrets) > 0 {
		desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets = imageScan.Spec.ImagePullSecrets

		// Mount docker config secrets as volumes for Syft to use
		for i, secretRef := range imageScan.Spec.ImagePullSecrets {
			volumeName := fmt.Sprintf("docker-config-%d", i)

			// Add volume
			desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes = append(
				desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: volumeName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretRef.Name,
							Items: []corev1.KeyToPath{
								{
									Key:  ".dockerconfigjson",
									Path: "config.json",
								},
							},
						},
					},
				},
			)

			// Add volume mount
			desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts = append(
				desiredCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      volumeName,
					MountPath: fmt.Sprintf("/docker-config/%d", i),
					ReadOnly:  true,
				},
			)
		}
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(imageScan, desiredCronJob, r.Scheme); err != nil {
		return nil, err
	}

	if errors.IsNotFound(err) {
		// Create the CronJob
		logger.Info("Creating CronJob", "name", cronJobName)
		if err := r.Create(ctx, desiredCronJob); err != nil {
			return nil, err
		}
		return desiredCronJob, nil
	} else if err != nil {
		return nil, err
	}

	// Update existing CronJob
	cronJob.Spec = desiredCronJob.Spec
	logger.Info("Updating CronJob", "name", cronJobName)
	if err := r.Update(ctx, cronJob); err != nil {
		return nil, err
	}

	return cronJob, nil
}

// handleDeletion handles the deletion of an ImageScan
func (r *ImageScanReconciler) handleDeletion(ctx context.Context, imageScan *invulnerablev1alpha1.ImageScan) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(imageScan, imageScanFinalizer) {
		// Delete the CronJob
		cronJobName := fmt.Sprintf("%s-scanner", imageScan.Name)
		cronJob := &batchv1.CronJob{}
		err := r.Get(ctx, types.NamespacedName{Name: cronJobName, Namespace: imageScan.Namespace}, cronJob)
		if err == nil {
			logger.Info("Deleting CronJob", "name", cronJobName)
			if err := r.Delete(ctx, cronJob); err != nil {
				return ctrl.Result{}, err
			}
		} else if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(imageScan, imageScanFinalizer)
		if err := r.Update(ctx, imageScan); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// setCondition sets a condition on the ImageScan status
func (r *ImageScanReconciler) setCondition(imageScan *invulnerablev1alpha1.ImageScan, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: imageScan.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
	}

	meta.SetStatusCondition(&imageScan.Status.Conditions, condition)
}

// buildEnvVars builds the environment variables for the scanner container
func buildEnvVars(imageScan *invulnerablev1alpha1.ImageScan, apiEndpoint, sbomFormat string) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  "SCAN_IMAGE",
			Value: imageScan.Spec.Image,
		},
		{
			Name:  "API_ENDPOINT",
			Value: apiEndpoint,
		},
		{
			Name:  "SBOM_FORMAT",
			Value: sbomFormat,
		},
		{
			Name:  "SYFT_CACHE_DIR",
			Value: "/tmp/syft/cache",
		},
		{
			Name:  "GRYPE_DB_CACHE_DIR",
			Value: "/tmp/syft/grype-db",
		},
	}

	// Set DOCKER_CONFIG for Syft to find registry credentials
	if len(imageScan.Spec.ImagePullSecrets) > 0 {
		env = append(env, corev1.EnvVar{
			Name:  "DOCKER_CONFIG",
			Value: "/docker-config/0",
		})
	}

	// Add webhook configuration if present and enabled
	if imageScan.Spec.Webhook != nil && imageScan.Spec.Webhook.Enabled {
		env = append(env,
			corev1.EnvVar{
				Name:  "WEBHOOK_URL",
				Value: imageScan.Spec.Webhook.URL,
			},
			corev1.EnvVar{
				Name:  "WEBHOOK_FORMAT",
				Value: imageScan.Spec.Webhook.Format,
			},
			corev1.EnvVar{
				Name:  "WEBHOOK_MIN_SEVERITY",
				Value: imageScan.Spec.Webhook.MinSeverity,
			},
		)
	}

	return env
}

// SetupWithManager sets up the controller with the Manager
func (r *ImageScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&invulnerablev1alpha1.ImageScan{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}

// ptr returns a pointer to the provided value
func ptr[T any](v T) *T {
	return &v
}

// resourceQuantityPtr parses a resource quantity string and returns a pointer to it
func resourceQuantityPtr(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}
