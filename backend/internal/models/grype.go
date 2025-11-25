package models

// GrypeResult represents the structure of Grype's JSON output
type GrypeResult struct {
	Matches        []GrypeMatch      `json:"matches"`
	Source         *GrypeSource      `json:"source,omitempty"`
	Descriptor     GrypeDescriptor   `json:"descriptor"`
	Distro         *GrypeDistro      `json:"distro,omitempty"`
}

type GrypeMatch struct {
	Vulnerability GrypeVulnerability `json:"vulnerability"`
	RelatedVulnerabilities []GrypeRelatedVuln `json:"relatedVulnerabilities,omitempty"`
	MatchDetails  []GrypeMatchDetail `json:"matchDetails"`
	Artifact      GrypeArtifact      `json:"artifact"`
}

type GrypeVulnerability struct {
	ID          string              `json:"id"`
	DataSource  string              `json:"dataSource,omitempty"`
	Namespace   string              `json:"namespace,omitempty"`
	Severity    string              `json:"severity"`
	URLs        []string            `json:"urls,omitempty"`
	Description string              `json:"description,omitempty"`
	Cvss        []GrypeCVSS         `json:"cvss,omitempty"`
	Fix         *GrypeFix           `json:"fix,omitempty"`
	Advisories  []GrypeAdvisory     `json:"advisories,omitempty"`
}

type GrypeRelatedVuln struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace,omitempty"`
}

type GrypeCVSS struct {
	Version string                 `json:"version"`
	Vector  string                 `json:"vector"`
	Metrics map[string]interface{} `json:"metrics"`
	VendorMetadata map[string]interface{} `json:"vendorMetadata,omitempty"`
}

type GrypeFix struct {
	Versions []string `json:"versions"`
	State    string   `json:"state,omitempty"`
}

type GrypeAdvisory struct {
	ID   string `json:"id"`
	Link string `json:"link,omitempty"`
}

type GrypeMatchDetail struct {
	Type       string                 `json:"type"`
	Matcher    string                 `json:"matcher"`
	SearchedBy map[string]interface{} `json:"searchedBy,omitempty"`
	Found      map[string]interface{} `json:"found,omitempty"`
}

type GrypeArtifact struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Type      string            `json:"type"`
	Locations []GrypeLocation   `json:"locations,omitempty"`
	Language  string            `json:"language,omitempty"`
	Licenses  []string          `json:"licenses,omitempty"`
	CPEs      []string          `json:"cpes,omitempty"`
	PURL      string            `json:"purl,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type GrypeLocation struct {
	Path string `json:"path"`
}

type GrypeSource struct {
	Type   string                 `json:"type"`
	Target map[string]interface{} `json:"target"`
}

type GrypeDescriptor struct {
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type GrypeDistro struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	IDLike  []string `json:"idLike,omitempty"`
}
