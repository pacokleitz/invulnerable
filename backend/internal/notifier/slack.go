package notifier

import "fmt"

type SlackPayload struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Color  string       `json:"color,omitempty"`
	Text   string       `json:"text,omitempty"`
	Fields []SlackField `json:"fields,omitempty"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func (n *Notifier) buildSlackPayload(payload NotificationPayload) SlackPayload {
	// Determine color based on severity
	color := n.getSeverityColor(payload.SeverityCounts)

	// Build summary text
	var summaryText string
	if payload.TotalVulns == 0 {
		summaryText = fmt.Sprintf("✅ No vulnerabilities found in `%s`", payload.Image)
	} else {
		summaryText = fmt.Sprintf("⚠️ Found %d vulnerabilities in `%s`", payload.TotalVulns, payload.Image)
	}

	fields := []SlackField{
		{Title: "Critical", Value: fmt.Sprintf("%d", payload.SeverityCounts.Critical), Short: true},
		{Title: "High", Value: fmt.Sprintf("%d", payload.SeverityCounts.High), Short: true},
		{Title: "Medium", Value: fmt.Sprintf("%d", payload.SeverityCounts.Medium), Short: true},
		{Title: "Low", Value: fmt.Sprintf("%d", payload.SeverityCounts.Low), Short: true},
	}

	if payload.ImageDigest != nil {
		fields = append(fields, SlackField{
			Title: "Digest",
			Value: *payload.ImageDigest,
			Short: false,
		})
	}

	// Add scan URL if available
	if payload.ScanURL != "" {
		fields = append(fields, SlackField{
			Title: "View Scan",
			Value: fmt.Sprintf("<%s|View full scan results>", payload.ScanURL),
			Short: false,
		})
	}

	return SlackPayload{
		Text: summaryText,
		Attachments: []SlackAttachment{
			{
				Color:  color,
				Text:   "Vulnerability Summary",
				Fields: fields,
			},
		},
	}
}

func (n *Notifier) getSeverityColor(counts SeverityCounts) string {
	if counts.Critical > 0 {
		return "danger" // Red
	}
	if counts.High > 0 {
		return "warning" // Orange
	}
	if counts.Medium > 0 {
		return "#ffcc00" // Yellow
	}
	return "good" // Green
}

func (n *Notifier) buildSlackStatusChangePayload(payload StatusChangeNotificationPayload) SlackPayload {
	// Determine color based on new status
	color := n.getStatusChangeColor(payload.NewStatus)

	// Build summary text
	summaryText := fmt.Sprintf("🔔 Vulnerability status changed: `%s` in `%s`",
		payload.CVEID, payload.ImageName)

	// Build fields
	fields := []SlackField{
		{Title: "CVE ID", Value: payload.CVEID, Short: true},
		{Title: "Severity", Value: payload.Severity, Short: true},
		{Title: "Package", Value: fmt.Sprintf("%s (%s)", payload.PackageName, payload.PackageVersion), Short: false},
		{Title: "Status Change", Value: fmt.Sprintf("%s → %s", payload.OldStatus, payload.NewStatus), Short: true},
		{Title: "Changed By", Value: payload.ChangedBy, Short: true},
	}

	// Add notes if present
	if payload.Notes != nil && *payload.Notes != "" {
		fields = append(fields, SlackField{
			Title: "Notes",
			Value: *payload.Notes,
			Short: false,
		})
	}

	// Add vulnerability URL if available
	if payload.VulnURL != "" {
		fields = append(fields, SlackField{
			Title: "View Details",
			Value: fmt.Sprintf("<%s|View vulnerability details>", payload.VulnURL),
			Short: false,
		})
	}

	return SlackPayload{
		Text: summaryText,
		Attachments: []SlackAttachment{
			{
				Color:  color,
				Text:   "Status Change Details",
				Fields: fields,
			},
		},
	}
}

func (n *Notifier) getStatusChangeColor(status string) string {
	switch status {
	case "fixed":
		return "good" // Green
	case "active":
		return "danger" // Red
	case "ignored":
		return "#9E9E9E" // Gray
	case "in_progress":
		return "warning" // Orange
	case "false_positive":
		return "#2196F3" // Blue
	default:
		return "#757575" // Default gray
	}
}
