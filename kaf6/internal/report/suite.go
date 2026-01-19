package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"kaf6/internal/engine"
)

type SuiteResult struct {
	RunID     string          `json:"run_id"`
	StartedAt time.Time       `json:"started_at"`
	Duration  time.Duration   `json:"duration"`
	Results   []engine.Result `json:"results"`
}

func WriteSuite(runID string, results []engine.Result, root string) (string, string, error) {
	if root == "" {
		root = "reports"
	}
	dir := filepath.Join(root, runID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	suite := SuiteResult{
		RunID:     runID,
		StartedAt: time.Now(),
		Duration:  0,
		Results:   results,
	}
	jsonPath := filepath.Join(dir, "suite.json")
	if err := writeSuiteJSON(jsonPath, suite); err != nil {
		return "", "", err
	}
	htmlPath := filepath.Join(dir, "suite.html")
	data := BuildReportData(results, fmt.Sprintf("KAF6 Suite %s", runID), runID)
	reportPath := filepath.Join(dir, "report.json")
	if err := WriteReportData(reportPath, data); err != nil {
		return "", "", err
	}
	if err := WriteHTMLFromData(htmlPath, data); err != nil {
		return "", "", err
	}
	return jsonPath, htmlPath, nil
}

func writeSuiteJSON(path string, suite SuiteResult) error {
	payload, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func writeSuiteHTML(path string, suite SuiteResult) error {
	data := BuildReportData(suite.Results, fmt.Sprintf("KAF6 Suite %s", suite.RunID), suite.RunID)
	return writeUnifiedHTML(path, data)
}
