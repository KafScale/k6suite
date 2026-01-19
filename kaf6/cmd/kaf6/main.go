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

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"

	"kaf6/internal/engine"
	"kaf6/internal/profile"
	"kaf6/internal/report"
	"kaf6/internal/scenario"
)

func main() {
	var reportDir string
	var suiteDir string
	flag.StringVar(&reportDir, "report-dir", "reports", "report output directory")
	flag.StringVar(&suiteDir, "suite", "", "suite directory with JSON scenarios")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "usage: kaf6 run <scenario.json> [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 run-suite <dir> [--suite dir] [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 select <dir> [--suite dir] [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 k6-select <dir> [--suite dir]")
		fmt.Fprintln(os.Stderr, "       kaf6 render-report <report.json> [--report-dir reports]")
		os.Exit(2)
	}
	switch flag.Arg(0) {
	case "run":
		path := flag.Arg(1)
		spec, err := scenario.Load(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load scenario: %v\n", err)
			os.Exit(1)
		}
		result, err := engine.Run(context.Background(), spec)
		if result != nil {
			jsonPath, htmlPath, repErr := report.Write(result, reportDir)
			if repErr != nil {
				fmt.Fprintf(os.Stderr, "write report: %v\n", repErr)
				os.Exit(1)
			}
			fmt.Printf("status: %s\nsummary: %s\nreport: %s\n", result.Status, jsonPath, htmlPath)
			if os.Getenv("KAF6_OPEN") == "1" {
				_ = exec.Command("open", htmlPath).Run()
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "run scenario: %v\n", err)
			os.Exit(1)
		}
	case "run-suite":
		dir := flag.Arg(1)
		if suiteDir != "" {
			dir = suiteDir
		}
		dir = strings.TrimSpace(dir)
		if strings.HasPrefix(dir, "--") || dir == "" {
			dir = "suite"
		}
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan suite: %v\n", err)
			os.Exit(1)
		}
		files = filterSuiteFiles(files)
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "no scenario files found in %s\n", dir)
			os.Exit(1)
		}
		sort.Strings(files)
		statusPath := filepath.Join(dir, "status.md")
		hashes, err := validateSuite(files, dir, statusPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate suite: %v\n", err)
			os.Exit(1)
		}
		if err := verifySuite(files, dir, statusPath, hashes); err != nil {
			fmt.Fprintf(os.Stderr, "suite integrity check failed: %v\n", err)
			os.Exit(1)
		}
		profileFile, profileIDs, profileSource, err := loadProfileIDs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load profiles: %v\n", err)
			os.Exit(1)
		}
		if len(profileIDs) == 0 {
			profileIDs = []string{""}
		}
		jsonPath, htmlPath, err := runSuite(dir, files, profileIDs, profileFile, profileSource, reportDir)
		if jsonPath != "" || htmlPath != "" {
			fmt.Printf("suite summary: %s\nsuite report: %s\n", jsonPath, htmlPath)
			if os.Getenv("KAF6_OPEN") == "1" {
				_ = exec.Command("open", htmlPath).Run()
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "run suite: %v\n", err)
			if os.Getenv("KAF6_ALLOW_FAIL") != "1" {
				os.Exit(1)
			}
		}
	case "select":
		dir := flag.Arg(1)
		if suiteDir != "" {
			dir = suiteDir
		}
		dir = strings.TrimSpace(dir)
		if strings.HasPrefix(dir, "--") || dir == "" {
			dir = "suite"
		}
		files, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan suite: %v\n", err)
			os.Exit(1)
		}
		files = filterSuiteFiles(files)
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "no scenario files found in %s\n", dir)
			os.Exit(1)
		}
		sort.Strings(files)
		statusPath := filepath.Join(dir, "status.md")
		hashes, err := validateSuite(files, dir, statusPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate suite: %v\n", err)
			os.Exit(1)
		}
		if err := verifySuite(files, dir, statusPath, hashes); err != nil {
			fmt.Fprintf(os.Stderr, "suite integrity check failed: %v\n", err)
			os.Exit(1)
		}
		profileFile, profileIDs, profileSource, err := loadProfileIDs(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load profiles: %v\n", err)
			os.Exit(1)
		}
		selectedFiles, err := selectScenarios(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select scenarios: %v\n", err)
			os.Exit(1)
		}
		selectedProfiles, err := selectProfiles(profileIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select profiles: %v\n", err)
			os.Exit(1)
		}
		jsonPath, htmlPath, err := runSuite(dir, selectedFiles, selectedProfiles, profileFile, profileSource, reportDir)
		if jsonPath != "" || htmlPath != "" {
			fmt.Printf("suite summary: %s\nsuite report: %s\n", jsonPath, htmlPath)
			if os.Getenv("KAF6_OPEN") == "1" {
				_ = exec.Command("open", htmlPath).Run()
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "run suite: %v\n", err)
			if os.Getenv("KAF6_ALLOW_FAIL") != "1" {
				os.Exit(1)
			}
		}
	case "render-report":
		path := flag.Arg(1)
		data, err := report.ReadReportData(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read report data: %v\n", err)
			os.Exit(1)
		}
		dir := filepath.Dir(path)
		if reportDir != "" && reportDir != "reports" {
			dir = reportDir
		}
		htmlPath := filepath.Join(dir, "report.html")
		if err := report.WriteHTMLFromData(htmlPath, data); err != nil {
			fmt.Fprintf(os.Stderr, "write report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("report: %s\n", htmlPath)
		if os.Getenv("KAF6_OPEN") == "1" {
			_ = exec.Command("open", htmlPath).Run()
		}
	case "k6-select":
		dir := flag.Arg(1)
		if suiteDir != "" {
			dir = suiteDir
		}
		dir = strings.TrimSpace(dir)
		if strings.HasPrefix(dir, "--") || dir == "" {
			dir = "tests/k6"
		}
		if !pathExists(dir) {
			alt := filepath.Join("..", "tests", "k6")
			if pathExists(alt) {
				dir = alt
			}
		}
		files, err := filepath.Glob(filepath.Join(dir, "*.js"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan k6 tests: %v\n", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Fprintf(os.Stderr, "no k6 scenario files found in %s\n", dir)
			os.Exit(1)
		}
		sort.Strings(files)
		selectedFiles, err := selectK6Scenarios(files)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select scenarios: %v\n", err)
			os.Exit(1)
		}
		profileIDs, err := loadK6Profiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "load profiles: %v\n", err)
			os.Exit(1)
		}
		selectedProfiles, err := selectProfiles(profileIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "select profiles: %v\n", err)
			os.Exit(1)
		}
		if err := runK6Suite(selectedFiles, selectedProfiles); err != nil {
			fmt.Fprintf(os.Stderr, "run k6 suite: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "usage: kaf6 run <scenario.json> [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 run-suite <dir> [--suite dir] [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 select <dir> [--suite dir] [--report-dir reports]")
		fmt.Fprintln(os.Stderr, "       kaf6 k6-select <dir> [--suite dir]")
		fmt.Fprintln(os.Stderr, "       kaf6 render-report <report.json> [--report-dir reports]")
		os.Exit(2)
	}
}

func validateSuite(files []string, dir string, statusPath string) (map[string]string, error) {
	hashes := make(map[string]string, len(files))
	for _, file := range files {
		if _, err := scenario.Load(file); err != nil {
			return nil, fmt.Errorf("validate %s: %w", file, err)
		}
		hash, err := hashFile(file)
		if err != nil {
			return nil, err
		}
		rel := file
		if relative, err := filepath.Rel(dir, file); err == nil {
			rel = relative
		}
		hashes[rel] = hash
	}
	if err := writeStatus(statusPath, hashes); err != nil {
		return nil, err
	}
	return hashes, nil
}

func verifySuite(files []string, dir string, statusPath string, hashes map[string]string) error {
	current, err := readStatus(statusPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		rel := file
		if relative, err := filepath.Rel(dir, file); err == nil {
			rel = relative
		}
		hash, err := hashFile(file)
		if err != nil {
			return err
		}
		expected, ok := current[rel]
		if !ok {
			return fmt.Errorf("missing hash for %s", rel)
		}
		if hash != expected {
			return fmt.Errorf("hash mismatch for %s", rel)
		}
	}
	return nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func writeStatus(path string, hashes map[string]string) error {
	lines := []string{
		"# kaf6 suite validation",
		fmt.Sprintf("validated_at: %s", time.Now().Format(time.RFC3339)),
		"",
	}
	keys := make([]string, 0, len(hashes))
	for file := range hashes {
		keys = append(keys, file)
	}
	sort.Strings(keys)
	for _, file := range keys {
		lines = append(lines, fmt.Sprintf("%s  %s", hashes[file], file))
	}
	lines = append(lines, "")
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

func readStatus(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	hashes := make(map[string]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "validated_at:") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hashes[parts[1]] = parts[0]
	}
	return hashes, nil
}

func filterSuiteFiles(files []string) []string {
	filtered := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "profiles.json") {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func runSuite(dir string, files []string, profileIDs []string, profileFile *profile.ProfileFile, profileSource string, reportDir string) (string, string, error) {
	runID := time.Now().Format("20060102-150405")
	results := make([]engine.Result, 0, len(files))
	if len(profileIDs) == 0 {
		profileIDs = []string{""}
	}
	var runErrs []string
	for _, profileID := range profileIDs {
		for _, file := range files {
			spec, err := scenario.Load(file)
			if err != nil {
				runErrs = append(runErrs, fmt.Sprintf("load scenario %s: %v", file, err))
				continue
			}
			if profileID != "" {
				if err := applyProfileOverride(spec, profileFile, profileID, profileSource); err != nil {
					runErrs = append(runErrs, fmt.Sprintf("apply profile %s: %v", profileID, err))
					continue
				}
			}
			result, err := engine.Run(context.Background(), spec)
			if result != nil {
				results = append(results, *result)
			}
			if err != nil {
				runErrs = append(runErrs, fmt.Sprintf("run scenario %s: %v", file, err))
			}
		}
	}
	jsonPath, htmlPath, repErr := report.WriteSuite(runID, results, reportDir)
	if repErr != nil {
		return jsonPath, htmlPath, repErr
	}
	if len(runErrs) > 0 {
		return jsonPath, htmlPath, fmt.Errorf(strings.Join(runErrs, "; "))
	}
	return jsonPath, htmlPath, nil
}

func loadProfileIDs(suiteDir string) (*profile.ProfileFile, []string, string, error) {
	primary := filepath.Join(suiteDir, "profiles.json")
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, "", err
	}
	fallback := filepath.Join(cwd, "config", "profiles.json")
	file, source, err := profile.LoadWithFallback(primary, fallback)
	if err != nil {
		return nil, nil, "", nil
	}
	ids := make([]string, 0, len(file.Profiles))
	for id := range file.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return file, ids, source, nil
}

func applyProfileOverride(spec *scenario.ScenarioFile, profileFile *profile.ProfileFile, profileID string, source string) error {
	if profileID == "" || profileFile == nil {
		return nil
	}
	resolved, err := profile.Resolve(profileFile, profileID)
	if err != nil {
		return err
	}
	spec.Profile = profileID
	spec.ProfileName = resolved.Name
	spec.ProfileDescription = resolved.Description
	spec.ProfileMetricsURL = resolved.MetricsURL
	spec.ProfileSource = source
	spec.Brokers = resolved.Brokers
	if spec.Scenarios.Metrics != nil && spec.Scenarios.Metrics.URL == "" {
		spec.Scenarios.Metrics.URL = resolved.MetricsURL
	}
	return nil
}

func selectScenarios(files []string) ([]string, error) {
	options := make([]string, 0, len(files)+1)
	options = append(options, "All scenarios")
	indexByName := make(map[string]string, len(files))
	for _, file := range files {
		name := filepath.Base(file)
		options = append(options, name)
		indexByName[name] = file
	}
	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select scenarios:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no scenarios selected")
	}
	for _, value := range selected {
		if value == "All scenarios" {
			return files, nil
		}
	}
	result := make([]string, 0, len(selected))
	for _, value := range selected {
		if file, ok := indexByName[value]; ok {
			result = append(result, file)
		}
	}
	sort.Strings(result)
	return result, nil
}

func selectProfiles(profileIDs []string) ([]string, error) {
	options := []string{"All profiles"}
	if len(profileIDs) == 0 {
		options = append(options, "Default")
	} else {
		options = append(options, profileIDs...)
	}
	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select profiles:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no profiles selected")
	}
	for _, value := range selected {
		if value == "All profiles" {
			if len(profileIDs) == 0 {
				return []string{""}, nil
			}
			return profileIDs, nil
		}
	}
	result := make([]string, 0, len(selected))
	for _, value := range selected {
		if value == "Default" {
			result = append(result, "")
			continue
		}
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func selectK6Scenarios(files []string) ([]string, error) {
	options := make([]string, 0, len(files)+1)
	options = append(options, "All scenarios")
	indexByName := make(map[string]string, len(files))
	for _, file := range files {
		name := filepath.Base(file)
		options = append(options, name)
		indexByName[name] = file
	}
	var selected []string
	prompt := &survey.MultiSelect{
		Message: "Select k6 scenarios:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("no scenarios selected")
	}
	for _, value := range selected {
		if value == "All scenarios" {
			return files, nil
		}
	}
	result := make([]string, 0, len(selected))
	for _, value := range selected {
		if file, ok := indexByName[value]; ok {
			result = append(result, file)
		}
	}
	sort.Strings(result)
	return result, nil
}

func loadK6Profiles() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	primary := filepath.Join(cwd, "config", "profiles.json")
	fallback := filepath.Join(cwd, "kaf6", "config", "profiles.json")
	file, _, err := profile.LoadWithFallback(primary, fallback)
	if err != nil {
		return []string{}, nil
	}
	ids := make([]string, 0, len(file.Profiles))
	for id := range file.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func runK6Suite(files []string, profileIDs []string) error {
	binary := resolveK6Binary()
	if len(profileIDs) == 0 {
		profileIDs = []string{""}
	}
	var errs []string
	for _, profileID := range profileIDs {
		for _, file := range files {
			cmd := exec.Command(binary, "run", file)
			cmd.Env = os.Environ()
			if profileID != "" {
				cmd.Env = append(cmd.Env, "K6_PROFILE="+profileID)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				errs = append(errs, fmt.Sprintf("%s (%s): %v", file, profileLabel(profileID), err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

func resolveK6Binary() string {
	if custom := os.Getenv("K6_BIN"); custom != "" {
		return custom
	}
	cwd, err := os.Getwd()
	if err == nil {
		local := filepath.Join(cwd, "k6")
		if _, err := os.Stat(local); err == nil {
			return local
		}
		parent := filepath.Join(cwd, "..", "k6")
		if _, err := os.Stat(parent); err == nil {
			return parent
		}
	}
	return "k6"
}

func profileLabel(profileID string) string {
	if profileID == "" {
		return "default"
	}
	return profileID
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
