package prompts

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAllPromptFilesAreLoaded(t *testing.T) {
	if err := LoadError(); err != nil {
		t.Fatalf("LoadError() = %v", err)
	}

	for section, keys := range promptFiles {
		for key := range keys {
			if prompt := Get(section, key); prompt == "" {
				t.Errorf("Get(%q, %q) returned an empty prompt", section, key)
			}
		}
	}
}

func TestSystemPromptYAMLIsAssembledFromPromptFiles(t *testing.T) {
	data, err := SystemPromptYAML()
	if err != nil {
		t.Fatalf("SystemPromptYAML() error = %v", err)
	}

	var config struct {
		System struct {
			Name string `yaml:"name"`
		} `yaml:"system"`
		Goals []string `yaml:"goals"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("SystemPromptYAML() returned invalid YAML: %v", err)
	}
	if config.System.Name != "Runa" {
		t.Errorf("system.name = %q, want %q", config.System.Name, "Runa")
	}
	if len(config.Goals) != 4 {
		t.Errorf("len(goals) = %d, want 4", len(config.Goals))
	}
}
