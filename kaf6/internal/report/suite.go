// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
