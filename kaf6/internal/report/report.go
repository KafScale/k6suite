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
