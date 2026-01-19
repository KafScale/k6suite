package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kaf6/internal/profile"
)

type ScenarioFile struct {
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	Profile            string             `json:"profile"`
	ProfileName        string             `json:"-"`
	ProfileDescription string             `json:"-"`
	ProfileSource      string             `json:"-"`
	ProfileMetricsURL  string             `json:"-"`
	Brokers            []string           `json:"brokers"`
	Topics             []TopicSpec        `json:"topics"`
	Scenarios          ScenarioCollection `json:"scenarios"`
	Checks             []CheckSpec        `json:"checks"`
}

type TopicSpec struct {
	Name       string `json:"name"`
	Partitions int32  `json:"partitions"`
	Recreate   bool   `json:"recreate"`
}

type ScenarioCollection struct {
	Producer *ProducerScenario `json:"producer"`
	Consumer *ConsumerScenario `json:"consumer"`
	Metrics  *MetricsScenario  `json:"metrics"`
}

type ProducerScenario struct {
	Type     string         `json:"type"`
	Clients  int            `json:"clients"`
	Messages int            `json:"messages"`
	RatePerS float64        `json:"rate_per_s"`
	Topic    string         `json:"topic"`
	Value    PayloadSpec    `json:"value"`
	Headers  map[string]any `json:"headers"`
}

type ConsumerScenario struct {
	Type    string         `json:"type"`
	Clients int            `json:"clients"`
	Topic   string         `json:"topic"`
	Group   GroupSpec      `json:"group"`
	Offset  string         `json:"offset"`
	Limit   int            `json:"limit"`
	Timeout string         `json:"timeout"`
	Headers map[string]any `json:"headers"`
}

type MetricsScenario struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type GroupSpec struct {
	ID string `json:"id"`
}

type PayloadSpec struct {
	JSON map[string]string `json:"json"`
}

type CheckSpec struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Metric   string `json:"metric"`
	Expected int    `json:"expected"`
}

func Load(path string) (*ScenarioFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec ScenarioFile
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, err
	}
	if err := applyProfile(path, &spec); err != nil {
		return nil, err
	}
	if len(spec.Brokers) == 0 {
		return nil, fmt.Errorf("brokers are required")
	}
	if spec.Scenarios.Producer == nil && spec.Scenarios.Consumer == nil && spec.Scenarios.Metrics == nil {
		return nil, fmt.Errorf("at least one scenario is required")
	}
	return &spec, nil
}

func applyProfile(path string, spec *ScenarioFile) error {
	needsProfile := spec.Profile != "" || len(spec.Brokers) == 0
	if spec.Scenarios.Metrics != nil && spec.Scenarios.Metrics.URL == "" {
		needsProfile = true
	}
	suiteDir := filepath.Dir(path)
	primary := filepath.Join(suiteDir, "profiles.json")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fallback := filepath.Join(cwd, "config", "profiles.json")
	file, source, err := profile.LoadWithFallback(primary, fallback)
	if err != nil {
		if needsProfile {
			return err
		}
		return nil
	}
	profileID := spec.Profile
	if profileID == "" {
		profileID = file.DefaultProfile
	}
	resolved, err := profile.Resolve(file, profileID)
	if err != nil {
		return err
	}
	spec.Profile = profileID
	spec.ProfileName = resolved.Name
	spec.ProfileDescription = resolved.Description
	spec.ProfileSource = source
	spec.ProfileMetricsURL = resolved.MetricsURL
	if len(spec.Brokers) == 0 {
		spec.Brokers = resolved.Brokers
	}
	if spec.Scenarios.Metrics != nil && spec.Scenarios.Metrics.URL == "" {
		spec.Scenarios.Metrics.URL = resolved.MetricsURL
	}
	return nil
}
