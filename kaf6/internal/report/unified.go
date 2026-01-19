package report

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"kaf6/internal/engine"
)

type ReportData struct {
	Title     string
	RunID     string
	StartedAt time.Time
	Duration  time.Duration
	Summary   ReportSummary
	Groups    []ReportGroup
}

type ReportSummary struct {
	Status       string
	Connectivity string
	Profiles     int
	Scenarios    int
	Failed       int
	Produced     int64
	Consumed     int64
	Errors       int64
}

type ReportGroup struct {
	ProfileID          string
	ProfileName        string
	ProfileDescription string
	ProfileMetricsURL  string
	ProfileSource      string
	Summary            ReportSummary
	Results            []engine.Result
}

func BuildReportData(results []engine.Result, title string, runID string) ReportData {
	data := ReportData{
		Title: title,
		RunID: runID,
	}
	if len(results) == 0 {
		data.Summary.Status = "n.a."
		data.Summary.Connectivity = "n.a."
		return data
	}
	start := results[0].StartedAt
	end := results[0].StartedAt.Add(results[0].Duration)
	if title == "" {
		title = "KAF6 Report"
	}
	if data.RunID == "" {
		data.RunID = results[0].RunID
	}

	grouped := make(map[string]*ReportGroup)
	for _, result := range results {
		if result.StartedAt.Before(start) {
			start = result.StartedAt
		}
		finish := result.StartedAt.Add(result.Duration)
		if finish.After(end) {
			end = finish
		}
		key := profileKey(result.Profile, result.ProfileName)
		group := grouped[key]
		if group == nil {
			group = &ReportGroup{
				ProfileID:          result.Profile,
				ProfileName:        result.ProfileName,
				ProfileDescription: result.ProfileDescription,
				ProfileMetricsURL:  result.ProfileMetricsURL,
				ProfileSource:      result.ProfileSource,
				Summary: ReportSummary{
					Status:       "pass",
					Connectivity: "ok",
				},
			}
			grouped[key] = group
		}
		group.Results = append(group.Results, result)
		updateSummary(&group.Summary, result)
	}

	groups := make([]ReportGroup, 0, len(grouped))
	for _, group := range grouped {
		sort.Slice(group.Results, func(i, j int) bool {
			return group.Results[i].Name < group.Results[j].Name
		})
		groups = append(groups, *group)
	}
	sort.Slice(groups, func(i, j int) bool {
		return profileKey(groups[i].ProfileID, groups[i].ProfileName) < profileKey(groups[j].ProfileID, groups[j].ProfileName)
	})

	data.StartedAt = start
	data.Duration = end.Sub(start)
	data.Groups = groups

	summary := ReportSummary{
		Status:       "pass",
		Connectivity: "ok",
		Profiles:     len(groups),
	}
	for _, group := range groups {
		if group.Summary.Status != "pass" {
			summary.Status = "fail"
		}
		if group.Summary.Connectivity != "ok" {
			summary.Connectivity = "failed"
		}
		summary.Scenarios += group.Summary.Scenarios
		summary.Failed += group.Summary.Failed
		summary.Produced += group.Summary.Produced
		summary.Consumed += group.Summary.Consumed
		summary.Errors += group.Summary.Errors
	}
	data.Summary = summary
	return data
}

func WriteReportData(path string, data ReportData) error {
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func ReadReportData(path string) (ReportData, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ReportData{}, err
	}
	var data ReportData
	if err := json.Unmarshal(raw, &data); err != nil {
		return ReportData{}, err
	}
	return normalizeReportData(data), nil
}

func WriteHTMLFromData(path string, data ReportData) error {
	return writeUnifiedHTML(path, normalizeReportData(data))
}

func updateSummary(summary *ReportSummary, result engine.Result) {
	summary.Scenarios++
	if result.Status != "pass" {
		summary.Status = "fail"
		summary.Failed++
	}
	if result.ConnectivityStatus != "ok" {
		summary.Connectivity = "failed"
	}
	summary.Produced += result.Produced
	summary.Consumed += result.Consumed
	summary.Errors += result.Errors
}

func profileKey(id string, name string) string {
	if id == "" && name == "" {
		return "default"
	}
	if name != "" && name != id {
		return id + " (" + name + ")"
	}
	if id != "" {
		return id
	}
	return name
}

func writeUnifiedHTML(path string, data ReportData) error {
	data = normalizeReportData(data)
	content := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<title>%s</title>
<style>
body{font-family:Arial,Helvetica,sans-serif;margin:24px;color:#111}
h1{margin:0 0 8px 0}
h2{margin:20px 0 8px 0}
table{border-collapse:collapse;width:100%%;margin-top:12px}
th,td{border:1px solid #ddd;padding:8px;text-align:left;vertical-align:top}
th{background:#f3f3f3}
.ok{color:#0a7f2e;font-weight:bold}
.bad{color:#c21a1a;font-weight:bold}
.meta{color:#555;margin-bottom:16px}
.card{border:1px solid #e1e1e1;border-radius:10px;padding:12px 14px;margin:12px 0;background:#fafafa}
.card.status{border-color:#cfe6d1;background:#eef8f0}
.card.status.fail{border-color:#f0c7c7;background:#fff0f0}
.card.error{border-color:#f0c7c7;background:#fff0f0}
.card h3{font-size:16px;margin:0 0 8px 0}
.card dl{margin:0;display:grid;grid-template-columns:max-content 1fr;gap:6px 12px}
.card dt{color:#444;font-weight:bold}
.card dd{margin:0;color:#111}
.na{color:#777}
.tabs{display:flex;flex-wrap:wrap;gap:8px;margin:16px 0}
.tab-btn{border:1px solid #ddd;border-radius:999px;background:#f6f6f6;padding:6px 12px;cursor:pointer}
.tab-btn.active{border-color:#111;background:#111;color:#fff}
.tab-panel{display:none;margin-top:8px}
.tab-panel.active{display:block}
</style>
</head>
<body>
<h1>%s</h1>
<div class="meta">Started: %s</div>
<div class="meta">Duration: %s</div>
%s
%s
<script>
const tabButtons = document.querySelectorAll("[data-tab-target]");
const tabPanels = document.querySelectorAll("[data-tab-panel]");
tabButtons.forEach((button) => {
  button.addEventListener("click", () => {
    const target = button.getAttribute("data-tab-target");
    tabButtons.forEach((btn) => btn.classList.remove("active"));
    tabPanels.forEach((panel) => panel.classList.remove("active"));
    button.classList.add("active");
    const panel = document.querySelector(target);
    if (panel) panel.classList.add("active");
  });
});
if (tabButtons.length > 0 && tabPanels.length > 0) {
  tabButtons[0].classList.add("active");
  tabPanels[0].classList.add("active");
}
</script>
</body>
</html>`,
		data.Title,
		data.Title,
		formatTime(data.StartedAt),
		formatDuration(data.Duration),
		renderSummaryCard(data.Summary),
		renderProfileSections(data.Groups),
	)
	return os.WriteFile(path, []byte(content), 0o644)
}

func renderSummaryCard(summary ReportSummary) string {
	statusText := displayOrNA(summary.Status)
	statusClass := "card status"
	if statusText != "pass" && statusText != "n.a." {
		statusClass = "card status fail"
	}
	resultClass := "ok"
	if statusText != "pass" && statusText != "n.a." {
		resultClass = "bad"
	}
	connectivityText := displayOrNA(summary.Connectivity)
	connectivityClass := "ok"
	if connectivityText != "ok" && connectivityText != "n.a." {
		connectivityClass = "bad"
	}
	out := fmt.Sprintf(`<div class="%s"><h3>Overall Status</h3><dl>`, statusClass)
	out += fmt.Sprintf(`<dt>Result</dt><dd class="%s">%s</dd>`, resultClass, statusText)
	out += fmt.Sprintf(`<dt>Connectivity</dt><dd class="%s">%s</dd>`, connectivityClass, connectivityText)
	out += fmt.Sprintf(`<dt>Profiles</dt><dd>%s</dd>`, formatCount(summary.Profiles, summaryKnown(summary)))
	out += fmt.Sprintf(`<dt>Scenarios</dt><dd>%s</dd>`, formatCount(summary.Scenarios, summaryKnown(summary)))
	out += fmt.Sprintf(`<dt>Failed</dt><dd>%s</dd>`, formatCount(summary.Failed, summaryKnown(summary)))
	out += fmt.Sprintf(`<dt>Produced</dt><dd>%s</dd>`, formatCount64(summary.Produced, summaryKnown(summary)))
	out += fmt.Sprintf(`<dt>Consumed</dt><dd>%s</dd>`, formatCount64(summary.Consumed, summaryKnown(summary)))
	out += fmt.Sprintf(`<dt>Errors</dt><dd>%s</dd>`, formatCount64(summary.Errors, summaryKnown(summary)))
	out += `</dl></div>`
	return out
}

func renderProfileSections(groups []ReportGroup) string {
	if len(groups) == 0 {
		return ""
	}
	var sections []string
	sections = append(sections, renderProfileSummaryTable(groups))
	sections = append(sections, renderProfileTabs(groups))
	return strings.Join(sections, "\n")
}

func renderProfileLabel(id string, name string) string {
	if id == "" && name == "" {
		return ""
	}
	if name != "" && name != id {
		if id != "" {
			return fmt.Sprintf("%s (%s)", id, name)
		}
		return name
	}
	return id
}

func renderProfileSummaryTable(groups []ReportGroup) string {
	rows := ""
	for _, group := range groups {
		label, icon, statusClass := statusBadge(group.Summary.Status)
		known := len(group.Results) > 0
		rows += fmt.Sprintf(`<tr><td>%s</td><td class="%s">%s %s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			renderProfileLabel(group.ProfileID, group.ProfileName),
			statusClass,
			icon,
			label,
			displayOrNA(group.Summary.Connectivity),
			formatCount(group.Summary.Scenarios, known),
			formatCount(group.Summary.Failed, known),
			formatCount64(group.Summary.Errors, known),
		)
	}
	return fmt.Sprintf(`<h2>Totals by Profile</h2>
<table>
  <tr><th>Profile</th><th>Status</th><th>Connectivity</th><th>Scenarios</th><th>Failed</th><th>Errors</th></tr>
  %s
</table>`, rows)
}

func renderProfileTabs(groups []ReportGroup) string {
	if len(groups) == 0 {
		return ""
	}
	var tabs []string
	var panels []string
	for idx, group := range groups {
		label := renderProfileLabel(group.ProfileID, group.ProfileName)
		if label == "" {
			label = fmt.Sprintf("Profile %d", idx+1)
		}
		tabID := fmt.Sprintf("tab-%d", idx+1)
		tabs = append(tabs, fmt.Sprintf(`<button class="tab-btn" data-tab-target="#%s">%s</button>`, tabID, label))
		panels = append(panels, fmt.Sprintf(`<div class="tab-panel" id="%s" data-tab-panel>%s</div>`, tabID, renderProfileSection(group)))
	}
	return fmt.Sprintf(`<div class="tabs">%s</div>%s`, strings.Join(tabs, ""), strings.Join(panels, "\n"))
}

func renderProfileSection(group ReportGroup) string {
	statusText := displayOrNA(group.Summary.Status)
	statusClass := "card status"
	if statusText != "pass" && statusText != "n.a." {
		statusClass = "card status fail"
	}
	resultClass := "ok"
	if statusText != "pass" && statusText != "n.a." {
		resultClass = "bad"
	}
	connectivityText := displayOrNA(group.Summary.Connectivity)
	connectivityClass := "ok"
	if connectivityText != "ok" && connectivityText != "n.a." {
		connectivityClass = "bad"
	}
	card := fmt.Sprintf(`<div class="%s"><h3>Profile Status</h3><dl>`, statusClass)
	card += fmt.Sprintf(`<dt>Result</dt><dd class="%s">%s</dd>`, resultClass, statusText)
	card += fmt.Sprintf(`<dt>Connectivity</dt><dd class="%s">%s</dd>`, connectivityClass, connectivityText)
	card += fmt.Sprintf(`<dt>Scenarios</dt><dd>%s</dd>`, formatCount(group.Summary.Scenarios, len(group.Results) > 0))
	card += fmt.Sprintf(`<dt>Failed</dt><dd>%s</dd>`, formatCount(group.Summary.Failed, len(group.Results) > 0))
	card += fmt.Sprintf(`<dt>Errors</dt><dd>%s</dd>`, formatCount64(group.Summary.Errors, len(group.Results) > 0))
	card += `</dl></div>`

	profileCard := `<div class="card"><h3>Profile</h3><dl>`
	profileCard += fmt.Sprintf(`<dt>Name</dt><dd>%s</dd>`, displayOrNA(renderProfileLabel(group.ProfileID, group.ProfileName)))
	profileCard += fmt.Sprintf(`<dt>Description</dt><dd>%s</dd>`, displayOrNA(group.ProfileDescription))
	profileCard += fmt.Sprintf(`<dt>Metrics</dt><dd>%s</dd>`, displayOrNA(group.ProfileMetricsURL))
	profileCard += fmt.Sprintf(`<dt>Source</dt><dd>%s</dd>`, displayOrNA(group.ProfileSource))
	profileCard += `</dl></div>`

	rows := ""
	for _, result := range group.Results {
		label, icon, statusClass := statusBadge(result.Status)
		issues := renderIssues(result)
		rows += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td class="%s">%s %s</td><td>%d</td><td>%d</td><td>%d</td><td>%s</td></tr>`,
			displayOrNA(result.Name),
			displayOrNA(result.Description),
			statusClass,
			icon,
			label,
			result.Produced,
			result.Consumed,
			result.Errors,
			issues,
		)
	}

	errorCards := renderProfileErrors(group)
	table := fmt.Sprintf(`<table>
  <tr><th>Name</th><th>Description</th><th>Status</th><th>Produced</th><th>Consumed</th><th>Errors</th><th>Issues</th></tr>
  %s
</table>`, rows)

	return fmt.Sprintf(`<h2>Profile: %s</h2>%s%s%s%s`, renderProfileLabel(group.ProfileID, group.ProfileName), card, profileCard, errorCards, table)
}

func renderIssues(result engine.Result) string {
	parts := []string{}
	if result.ConnectivityError != "" {
		parts = append(parts, fmt.Sprintf("connectivity: %s", result.ConnectivityError))
	}
	if result.RunError != "" {
		parts = append(parts, fmt.Sprintf("run: %s", result.RunError))
	}
	if len(parts) == 0 {
		return "n.a."
	}
	return strings.Join(parts, "<br/>")
}

func renderProfileErrors(group ReportGroup) string {
	connectivityRows := ""
	runRows := ""
	for _, result := range group.Results {
		if result.ConnectivityError != "" {
			connectivityRows += fmt.Sprintf(`<li><strong>%s</strong>: %s</li>`, result.Name, result.ConnectivityError)
		}
		if result.RunError != "" {
			runRows += fmt.Sprintf(`<li><strong>%s</strong>: %s</li>`, result.Name, result.RunError)
		}
	}
	sections := ""
	if connectivityRows != "" {
		sections += fmt.Sprintf(`<div class="card error"><h3>Connectivity Errors</h3><ul>%s</ul></div>`, connectivityRows)
	}
	if runRows != "" {
		sections += fmt.Sprintf(`<div class="card error"><h3>Run Errors</h3><ul>%s</ul></div>`, runRows)
	}
	return sections
}

func normalizeReportData(data ReportData) ReportData {
	if data.Title == "" {
		data.Title = "KAF6 Report"
	}
	if data.Summary.Status == "" {
		data.Summary.Status = "n.a."
	}
	if data.Summary.Connectivity == "" {
		data.Summary.Connectivity = "n.a."
	}
	for i := range data.Groups {
		group := &data.Groups[i]
		if group.Summary.Status == "" {
			group.Summary.Status = "n.a."
		}
		if group.Summary.Connectivity == "" {
			group.Summary.Connectivity = "n.a."
		}
		for j := range group.Results {
			if group.Results[j].Status == "" {
				group.Results[j].Status = "n.a."
			}
			if group.Results[j].ConnectivityStatus == "" {
				group.Results[j].ConnectivityStatus = "n.a."
			}
		}
	}
	return data
}

func displayOrNA(value string) string {
	if strings.TrimSpace(value) == "" {
		return "n.a."
	}
	return value
}

func summaryKnown(summary ReportSummary) bool {
	return summary.Profiles > 0 || summary.Scenarios > 0 || summary.Failed > 0 || summary.Produced > 0 || summary.Consumed > 0 || summary.Errors > 0
}

func formatCount(value int, known bool) string {
	if !known {
		return "n.a."
	}
	return fmt.Sprintf("%d", value)
}

func formatCount64(value int64, known bool) string {
	if !known {
		return "n.a."
	}
	return fmt.Sprintf("%d", value)
}

func statusBadge(status string) (string, string, string) {
	text := displayOrNA(status)
	if text == "n.a." {
		return "n.a.", "&#8212;", "na"
	}
	if text == "pass" {
		return "pass", "&#x2705;", "ok"
	}
	return "fail", "&#x274C;", "bad"
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return "n.a."
	}
	return value.Format(time.RFC3339)
}

func formatDuration(value time.Duration) string {
	if value == 0 {
		return "n.a."
	}
	return value.String()
}
