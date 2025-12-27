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

	// Schedule in Cron format for when to run the scan
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:example="0 2 * * *"
	Schedule string `json:"schedule"`

	// TimeZone for the CronJob schedule (e.g., "America/New_York", "UTC")
	// If not specified, defaults to the system timezone
	// +kubebuilder:validation:Optional
	TimeZone *string `json:"timeZone,omitempty"`

	// SBOMFormat specifies the SBOM format to use (cyclonedx, spdx, etc.)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="cyclonedx"
	// +kubebuilder:validation:Enum=cyclonedx;spdx
	SBOMFormat string `json:"sbomFormat,omitempty"`

	// Suspend tells the controller to suspend subsequent scans
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

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

	// LastScheduledTime is the last time a job was scheduled
	// +kubebuilder:validation:Optional
	LastScheduledTime *metav1.Time `json:"lastScheduledTime,omitempty"`

	// LastSuccessfulTime is the last time a scan completed successfully
	// +kubebuilder:validation:Optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Conditions represent the latest available observations of the ImageScan's state
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation most recently observed by the controller
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=imgscan;imgscans
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Suspend",type=boolean,JSONPath=`.spec.suspend`
// +kubebuilder:printcolumn:name="CronJob",type=string,JSONPath=`.status.cronJobName`
// +kubebuilder:printcolumn:name="Last Scheduled",type=date,JSONPath=`.status.lastScheduledTime`
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
