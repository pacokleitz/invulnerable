package notifier

import "fmt"

type TeamsPayload struct {
	Type            string         `json:"@type"`
	Context         string         `json:"@context"`
	Summary         string         `json:"summary"`
	ThemeColor      string         `json:"themeColor"`
	Title           string         `json:"title"`
	Sections        []TeamsSection `json:"sections"`
	PotentialAction []TeamsAction  `json:"potentialAction,omitempty"`
}

type TeamsAction struct {
	Type    string        `json:"@type"`
	Name    string        `json:"name"`
	Targets []TeamsTarget `json:"targets,omitempty"`
}

type TeamsTarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

type TeamsSection struct {
	ActivityTitle    string      `json:"activityTitle,omitempty"`
	ActivitySubtitle string      `json:"activitySubtitle,omitempty"`
	Facts            []TeamsFact `json:"facts,omitempty"`
	Text             string      `json:"text,omitempty"`
}

type TeamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (n *Notifier) buildTeamsPayload(payload NotificationPayload) TeamsPayload {
	color := n.getTeamsColor(payload.SeverityCounts)

	var title, summary string
	if payload.TotalVulns == 0 {
		title = fmt.Sprintf("✅ Image Scan Passed: %s", payload.Image)
		summary = "No vulnerabilities found"
	} else {
		title = fmt.Sprintf("Image Scan Results: %s", payload.Image)
		summary = fmt.Sprintf("Found %d vulnerabilities", payload.TotalVulns)
	}

	facts := []TeamsFact{
		{Name: "Critical", Value: fmt.Sprintf("%d", payload.SeverityCounts.Critical)},
		{Name: "High", Value: fmt.Sprintf("%d", payload.SeverityCounts.High)},
		{Name: "Medium", Value: fmt.Sprintf("%d", payload.SeverityCounts.Medium)},
		{Name: "Low", Value: fmt.Sprintf("%d", payload.SeverityCounts.Low)},
		{Name: "Total Vulnerabilities", Value: fmt.Sprintf("%d", payload.TotalVulns)},
	}

	if payload.ImageDigest != nil {
		facts = append(facts, TeamsFact{
			Name:  "Image Digest",
			Value: *payload.ImageDigest,
		})
	}

	teamsPayload := TeamsPayload{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		Summary:    summary,
		ThemeColor: color,
		Title:      title,
		Sections: []TeamsSection{
			{
				ActivityTitle: "Vulnerability Summary",
				Facts:         facts,
			},
		},
	}

	// Add action button if scan URL is available
	if payload.ScanURL != "" {
		teamsPayload.PotentialAction = []TeamsAction{
			{
				Type: "OpenUri",
				Name: "View Scan Results",
				Targets: []TeamsTarget{
					{
						OS:  "default",
						URI: payload.ScanURL,
					},
				},
			},
		}
	}

	return teamsPayload
}

func (n *Notifier) getTeamsColor(counts SeverityCounts) string {
	if counts.Critical > 0 {
		return "FF0000" // Red
	}
	if counts.High > 0 {
		return "FFA500" // Orange
	}
	if counts.Medium > 0 {
		return "FFCC00" // Yellow
	}
	return "00FF00" // Green
}

func (n *Notifier) buildTeamsStatusChangePayload(payload StatusChangeNotificationPayload) TeamsPayload {
	color := n.getTeamsStatusChangeColor(payload.NewStatus)

	title := fmt.Sprintf("🔔 Vulnerability Status Changed: %s", payload.CVEID)
	summary := fmt.Sprintf("Status changed from %s to %s", payload.OldStatus, payload.NewStatus)

	facts := []TeamsFact{
		{Name: "CVE ID", Value: payload.CVEID},
		{Name: "Severity", Value: payload.Severity},
		{Name: "Package", Value: fmt.Sprintf("%s (%s)", payload.PackageName, payload.PackageVersion)},
		{Name: "Image", Value: payload.ImageName},
		{Name: "Status Change", Value: fmt.Sprintf("%s → %s", payload.OldStatus, payload.NewStatus)},
		{Name: "Changed By", Value: payload.ChangedBy},
	}

	// Add notes if present
	if payload.Notes != nil && *payload.Notes != "" {
		facts = append(facts, TeamsFact{
			Name:  "Notes",
			Value: *payload.Notes,
		})
	}

	teamsPayload := TeamsPayload{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		Summary:    summary,
		ThemeColor: color,
		Title:      title,
		Sections: []TeamsSection{
			{
				ActivityTitle: "Status Change Details",
				Facts:         facts,
			},
		},
	}

	// Add action button if vulnerability URL is available
	if payload.VulnURL != "" {
		teamsPayload.PotentialAction = []TeamsAction{
			{
				Type: "OpenUri",
				Name: "View Vulnerability Details",
				Targets: []TeamsTarget{
					{
						OS:  "default",
						URI: payload.VulnURL,
					},
				},
			},
		}
	}

	return teamsPayload
}

func (n *Notifier) getTeamsStatusChangeColor(status string) string {
	switch status {
	case "fixed":
		return "00FF00" // Green
	case "active":
		return "FF0000" // Red
	case "ignored":
		return "9E9E9E" // Gray
	case "in_progress":
		return "FFA500" // Orange
	case "false_positive":
		return "2196F3" // Blue
	default:
		return "757575" // Default gray
	}
}
