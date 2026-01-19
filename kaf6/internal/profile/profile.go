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

package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProfileFile struct {
	DefaultProfile string             `json:"default_profile"`
	Profiles       map[string]Profile `json:"profiles"`
}

type Profile struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Brokers     []string `json:"brokers"`
	MetricsURL  string   `json:"metrics_url"`
}

func Load() (*ProfileFile, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(cwd, "config", "profiles.json")
	return LoadWithFallback(path, "")
}

func LoadWithFallback(primary string, fallback string) (*ProfileFile, string, error) {
	path := primary
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			path = ""
		}
	}
	if path == "" && fallback != "" {
		if _, err := os.Stat(fallback); err == nil {
			path = fallback
		}
	}
	if path == "" {
		if primary != "" {
			return nil, "", fmt.Errorf("profiles.json not found at %s", primary)
		}
		return nil, "", fmt.Errorf("profiles.json not found")
	}
	file, err := loadFromPath(path)
	if err != nil {
		return nil, path, err
	}
	return file, path, nil
}

func loadFromPath(path string) (*ProfileFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file ProfileFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if file.DefaultProfile == "" {
		return nil, fmt.Errorf("default_profile is required")
	}
	return &file, nil
}

func Resolve(file *ProfileFile, profileName string) (Profile, error) {
	name := profileName
	if name == "" {
		name = file.DefaultProfile
	}
	profile, ok := file.Profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("unknown profile: %s", name)
	}
	if len(profile.Brokers) == 0 {
		return Profile{}, fmt.Errorf("profile %s missing brokers", name)
	}
	return profile, nil
}
