package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageScanSpec defines the desired state of ImageScan
type ImageScanSpec struct {
	// Image is the container image to scan (e.g., "nginx:latest", "myregistry.io/app:1.0.0")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Schedule configures time-based scanning via CronJob
	// If disabled or not specified, only registry polling will trigger scans
	// +kubebuilder:validation:Optional
	Schedule *ScheduleConfig `json:"schedule,omitempty"`

	// TimeZone for the CronJob schedule (e.g., "America/New_York", "UTC")
	// If not specified, defaults to the system timezone
	// +kubebuilder:validation:Optional
	TimeZone *string `json:"timeZone,omitempty"`

	// SBOMFormat specifies the SBOM format to use (cyclonedx, spdx, etc.)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="cyclonedx"
	// +kubebuilder:validation:Enum=cyclonedx;spdx
	SBOMFormat string `json:"sbomFormat,omitempty"`

	// SuccessfulJobsHistoryLimit is the number of successful jobs to retain
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// FailedJobsHistoryLimit is the number of failed jobs to retain
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=3
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`

	// Resources defines the resource requirements for the scanner job
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// WorkspaceSize defines the size of the temporary workspace for image extraction
	// This should be larger than the largest image you plan to scan
	// Default: 10Gi (suitable for most images, increase for larger images)
	// WARNING: Multiple ImageScans can run concurrently and consume node disk space
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="10Gi"
	WorkspaceSize string `json:"workspaceSize,omitempty"`

	// APIEndpoint is the Invulnerable backend API endpoint
	// If not specified, it will be auto-detected from the service
	// +kubebuilder:validation:Optional
	APIEndpoint string `json:"apiEndpoint,omitempty"`

	// Scanner image configuration
	// +kubebuilder:validation:Optional
	ScannerImage *ScannerImageSpec `json:"scannerImage,omitempty"`

	// Webhooks configuration for multiple notification types
	// +kubebuilder:validation:Optional
	Webhooks *WebhooksConfig `json:"webhooks,omitempty"`

	// ImagePullSecrets is an optional list of references to secrets in the same namespace
	// to use for pulling the container image from private registries.
	// These secrets should be of type kubernetes.io/dockerconfigjson.
	// See https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod
	// +kubebuilder:validation:Optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// OnlyFixable specifies whether to only report vulnerabilities with available fixes.
	// When true, Grype will skip vulnerabilities that have no fix available.
	// Default: false (report all vulnerabilities)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	OnlyFixable bool `json:"onlyFixable,omitempty"`

	// SLA defines Service Level Agreement for vulnerability remediation in days per severity.
	// This configuration is stored with each scan for compliance tracking.
	// If not specified, default SLA values are used: Critical=7, High=30, Medium=90, Low=180
	// +kubebuilder:validation:Optional
	SLA *SLAConfig `json:"sla,omitempty"`

	// RegistryPolling configures automatic scanning when image updates are detected in the registry
	// This is a hybrid approach - both scheduled CronJobs and registry-triggered scans can coexist
	// When enabled, the controller periodically checks the registry for digest changes and triggers immediate scans
	// +kubebuilder:validation:Optional
	RegistryPolling *RegistryPollingConfig `json:"registryPolling,omitempty"`
}

// SLAConfig defines remediation SLA in days for each severity level
type SLAConfig struct {
	// Critical severity SLA in days
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=7
	Critical int `json:"critical,omitempty"`

	// High severity SLA in days
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=30
	High int `json:"high,omitempty"`

	// Medium severity SLA in days
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=90
	Medium int `json:"medium,omitempty"`

	// Low severity SLA in days
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=180
	Low int `json:"low,omitempty"`
}

// ScheduleConfig defines time-based scanning configuration
type ScheduleConfig struct {
	// Enabled determines if scheduled scanning via CronJob is active
	// When false, only registry polling (if enabled) will trigger scans
	// Default: true
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// Cron schedule in cron format for when to run the scan
	// Required if enabled is true
	// Example: "0 2 * * *" (daily at 2 AM)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:example="0 2 * * *"
	Cron string `json:"cron,omitempty"`

	// Suspend temporarily pauses scheduled scans without disabling the schedule
	// When true, the CronJob will not trigger new scans, but registry polling (if enabled) continues
	// Default: false
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`
}

// RegistryPollingConfig defines configuration for automatic registry monitoring
// When enabled, the controller periodically checks the registry for image updates
// and triggers scans when the image digest changes
type RegistryPollingConfig struct {
	// Enabled determines if registry polling is active
	// When enabled, the controller will periodically check the registry for new image versions
	// and automatically trigger scans when the image digest changes
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Interval is how often to check the registry for updates
	// Must be at least 1 minute to prevent API rate limiting
	// Default: 5m (5 minutes)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="5m"
	Interval *metav1.Duration `json:"interval,omitempty"`
}

// WebhooksConfig defines configuration for multiple webhook notification types
// Both scanCompletion and statusChange webhooks share the same URL and format
type WebhooksConfig struct {
	// URL is the webhook endpoint URL used for all notification types
	// Either URL or SecretRef must be specified, but not both
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^https?://.*$`
	URL string `json:"url,omitempty"`

	// SecretRef references a Secret containing the webhook URL
	// The Secret must be in the same namespace as the ImageScan
	// Either URL or SecretRef must be specified, but not both
	// +kubebuilder:validation:Optional
	SecretRef *corev1.SecretKeySelector `json:"secretRef,omitempty"`

	// Format specifies the webhook payload format (slack, teams)
	// This format is used for all notification types
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=slack;teams
	// +kubebuilder:default="slack"
	Format string `json:"format,omitempty"`

	// ScanCompletion configures notifications sent after each scan completes
	// +kubebuilder:validation:Optional
	ScanCompletion *ScanCompletionWebhookConfig `json:"scanCompletion,omitempty"`

	// StatusChange configures notifications sent when vulnerability statuses are changed
	// +kubebuilder:validation:Optional
	StatusChange *StatusChangeWebhookConfig `json:"statusChange,omitempty"`
}

// ScanCompletionWebhookConfig defines webhook settings for scan completion notifications
type ScanCompletionWebhookConfig struct {
	// Enabled allows temporarily disabling scan completion notifications
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// MinSeverity is the minimum severity level to trigger notifications
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Critical;High;Medium;Low;Negligible
	// +kubebuilder:default="High"
	MinSeverity string `json:"minSeverity,omitempty"`

	// OnlyFixable specifies whether to only send notifications for vulnerabilities with available fixes.
	// When true, unfixable vulnerabilities will not trigger scan completion webhooks.
	// This is independent of the ImageScan's OnlyFixable setting - you can scan all CVEs but only notify for fixable ones.
	// Default: true (only notify for vulnerabilities with fixes)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	OnlyFixable bool `json:"onlyFixable,omitempty"`
}

// StatusChangeWebhookConfig defines webhook settings for vulnerability status change notifications
type StatusChangeWebhookConfig struct {
	// Enabled allows enabling/disabling status change notifications
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// MinSeverity is the minimum severity to notify about (Critical, High, Medium, Low, Negligible)
	// Only vulnerabilities at or above this severity will trigger notifications
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Critical;High;Medium;Low;Negligible
	// +kubebuilder:default="High"
	MinSeverity string `json:"minSeverity,omitempty"`

	// StatusTransitions filters which status changes trigger notifications
	// Format: "old_status→new_status" (e.g., "active→fixed", "active→ignored")
	// Empty list means notify on all transitions
	// Example: ["active→fixed", "active→ignored", "in_progress→fixed"]
	// +kubebuilder:validation:Optional
	StatusTransitions []string `json:"statusTransitions,omitempty"`

	// IncludeNoteChanges determines if note/comment additions trigger notifications
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	IncludeNoteChanges bool `json:"includeNoteChanges,omitempty"`

	// OnlyFixable specifies whether to only send notifications for vulnerabilities with available fixes.
	// When true, unfixable vulnerabilities will not trigger status change webhooks.
	// This is independent of the ImageScan's OnlyFixable setting - you can scan all CVEs but only notify for fixable ones.
	// Default: true (only notify for vulnerabilities with fixes)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	OnlyFixable bool `json:"onlyFixable,omitempty"`
}

// ScannerImageSpec defines the scanner container image configuration
type ScannerImageSpec struct {
	// Repository is the image repository
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="invulnerable-scanner"
	Repository string `json:"repository,omitempty"`

	// Tag is the image tag
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="latest"
	Tag string `json:"tag,omitempty"`

	// PullPolicy is the image pull policy
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +kubebuilder:default="IfNotPresent"
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

// ImageScanStatus defines the observed state of ImageScan
type ImageScanStatus struct {
	// CronJobName is the name of the managed CronJob
	// +kubebuilder:validation:Optional
	CronJobName string `json:"cronJobName,omitempty"`

	// LastSuccessfulTime is the last time a scan completed successfully
	// +kubebuilder:validation:Optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Conditions represent the latest available observations of the ImageScan's state
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation most recently observed by the controller
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// LastCheckedDigest is the image digest from the last registry check
	// Used to detect when a new image version has been pushed with the same tag
	// +kubebuilder:validation:Optional
	LastCheckedDigest string `json:"lastCheckedDigest,omitempty"`

	// LastRegistryCheckTime is when we last polled the registry for image updates
	// +kubebuilder:validation:Optional
	LastRegistryCheckTime *metav1.Time `json:"lastRegistryCheckTime,omitempty"`

	// NextRegistryCheckTime is when the next registry poll is scheduled
	// +kubebuilder:validation:Optional
	NextRegistryCheckTime *metav1.Time `json:"nextRegistryCheckTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=imgscan;imgscans
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Suspend",type=boolean,JSONPath=`.spec.suspend`
// +kubebuilder:printcolumn:name="CronJob",type=string,JSONPath=`.status.cronJobName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ImageScan is the Schema for the imagescans API
type ImageScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageScanSpec   `json:"spec,omitempty"`
	Status ImageScanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ImageScanList contains a list of ImageScan
type ImageScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImageScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImageScan{}, &ImageScanList{})
}
