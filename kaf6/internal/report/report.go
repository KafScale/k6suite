package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kaf6/internal/engine"
)

func Write(run *engine.Result, root string) (string, string, error) {
	if root == "" {
		root = "reports"
	}
	dir := filepath.Join(root, run.RunID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	jsonPath := filepath.Join(dir, "summary.json")
	if err := writeJSON(jsonPath, run); err != nil {
		return "", "", err
	}
	htmlPath := filepath.Join(dir, "report.html")
	data := BuildReportData([]engine.Result{*run}, fmt.Sprintf("KAF6 Run %s", run.RunID), run.RunID)
	reportPath := filepath.Join(dir, "report.json")
	if err := WriteReportData(reportPath, data); err != nil {
		return "", "", err
	}
	if err := WriteHTMLFromData(htmlPath, data); err != nil {
		return "", "", err
	}
	return jsonPath, htmlPath, nil
}

func writeJSON(path string, run *engine.Result) error {
	payload, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}
