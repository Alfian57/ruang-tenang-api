// Package prompts memusatkan seluruh prompt AI yang dipakai backend.
//
// Semua berkas prompt di folder ini di-embed ke dalam binary memakai go:embed,
// sehingga prompt tetap tersentralisasi di satu tempat (mudah diubah) sekaligus
// selalu tersedia di runtime tanpa bergantung pada working directory atau
// menyalin folder ke image Docker.
//
// Prompt dikelompokkan berdasarkan fitur. Setiap prompt operasional berada di
// file .txt sendiri, sedangkan konfigurasi system prompt chat tetap memakai
// format YAML agar struktur dan tipe datanya mudah dibaca oleh pemanggil.
package prompts

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed chat moderation journal wellness
var files embed.FS

// promptFiles memetakan section -> key -> path file prompt.
var promptFiles = map[string]map[string]string{
	"chat": {
		"transcribe":     "chat/transcribe.txt",
		"title":          "chat/title.txt",
		"summary":        "chat/summary.txt",
		"default_system": "chat/default_system.txt",
	},
	"moderation": {
		"article": "moderation/article.txt",
		"trigger": "moderation/trigger.txt",
		"forum":   "moderation/forum.txt",
	},
	"journal": {
		"single_summary":  "journal/single_summary.txt",
		"context_summary": "journal/context_summary.txt",
		"weekly_summary":  "journal/weekly_summary.txt",
	},
	"wellness": {
		"weekly_narrative": "wellness/weekly_narrative.txt",
	},
}

// promptsData memetakan section -> key -> teks prompt.
var (
	promptsData map[string]map[string]string
	loadOnce    sync.Once
	loadErr     error
)

func load() {
	loadOnce.Do(func() {
		promptsData = make(map[string]map[string]string, len(promptFiles))
		for section, keys := range promptFiles {
			promptsData[section] = make(map[string]string, len(keys))
			for key, path := range keys {
				data, err := files.ReadFile(path)
				if err != nil {
					loadErr = fmt.Errorf("prompts: gagal membaca %s: %w", path, err)
					return
				}
				promptsData[section][key] = string(data)
			}
		}
	})
}

// Get mengembalikan prompt mentah pada section/key. Mengembalikan string kosong
// bila tidak ditemukan (pemanggil sebaiknya menyediakan fallback).
func Get(section, key string) string {
	load()
	if loadErr != nil {
		return ""
	}
	if sec, ok := promptsData[section]; ok {
		return sec[key]
	}
	return ""
}

// Format mengambil prompt pada section/key lalu menerapkannya sebagai template
// fmt.Sprintf dengan args. Bila prompt tidak ditemukan, mengembalikan string
// kosong sehingga pemanggil dapat memakai fallback.
func Format(section, key string, args ...any) string {
	tmpl := Get(section, key)
	if tmpl == "" {
		return ""
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

type systemPromptFileConfig struct {
	System struct {
		Name    string `yaml:"name"`
		Context string `yaml:"context"`
		Persona string `yaml:"persona"`
	} `yaml:"system"`
	Goals        []string `yaml:"goals"`
	Instructions string   `yaml:"instructions"`
	Restrictions struct {
		AllowedTopics   []string `yaml:"allowed_topics"`
		ForbiddenTopics []string `yaml:"forbidden_topics"`
		Rejection       string   `yaml:"rejection_response"`
	} `yaml:"restrictions"`
	Security struct {
		IgnorePromptInjection bool     `yaml:"ignore_prompt_injection"`
		Rules                 []string `yaml:"rules"`
		InjectionResponse     string   `yaml:"injection_response"`
	} `yaml:"security"`
	CrisisHandling struct {
		Description            string `yaml:"description"`
		SelfHarmResponse       string `yaml:"self_harm_response"`
		ProfessionalDisclaimer string `yaml:"professional_disclaimer"`
	} `yaml:"crisis_handling"`
}

func readSystemText(path string) (string, error) {
	data, err := files.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("prompts: gagal membaca %s: %w", path, err)
	}
	return string(data), nil
}

func readSystemList(path string) ([]string, error) {
	text, err := readSystemText(path)
	if err != nil {
		return nil, err
	}

	var values []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			values = append(values, line)
		}
	}
	return values, nil
}

func readSystemSingleLine(path string) (string, error) {
	text, err := readSystemText(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(text), nil
}

// SystemPromptYAML merakit seluruh file teks system prompt chat menjadi YAML
// agar struktur yang dipakai pemanggil tetap kompatibel.
func SystemPromptYAML() ([]byte, error) {
	var config systemPromptFileConfig
	var err error
	if config.System.Name, err = readSystemSingleLine("chat/system/name.txt"); err != nil {
		return nil, err
	}
	if config.System.Context, err = readSystemSingleLine("chat/system/context.txt"); err != nil {
		return nil, err
	}
	if config.System.Persona, err = readSystemSingleLine("chat/system/persona.txt"); err != nil {
		return nil, err
	}
	if config.Goals, err = readSystemList("chat/system/goals.txt"); err != nil {
		return nil, err
	}
	if config.Instructions, err = readSystemText("chat/system/instructions.txt"); err != nil {
		return nil, err
	}
	if config.Restrictions.AllowedTopics, err = readSystemList("chat/system/restrictions/allowed_topics.txt"); err != nil {
		return nil, err
	}
	if config.Restrictions.ForbiddenTopics, err = readSystemList("chat/system/restrictions/forbidden_topics.txt"); err != nil {
		return nil, err
	}
	if config.Restrictions.Rejection, err = readSystemText("chat/system/restrictions/rejection_response.txt"); err != nil {
		return nil, err
	}
	config.Security.IgnorePromptInjection = true
	if config.Security.Rules, err = readSystemList("chat/system/security/rules.txt"); err != nil {
		return nil, err
	}
	if config.Security.InjectionResponse, err = readSystemText("chat/system/security/injection_response.txt"); err != nil {
		return nil, err
	}
	if config.CrisisHandling.Description, err = readSystemSingleLine("chat/system/crisis_handling/description.txt"); err != nil {
		return nil, err
	}
	if config.CrisisHandling.SelfHarmResponse, err = readSystemText("chat/system/crisis_handling/self_harm_response.txt"); err != nil {
		return nil, err
	}
	if config.CrisisHandling.ProfessionalDisclaimer, err = readSystemText("chat/system/crisis_handling/professional_disclaimer.txt"); err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("prompts: gagal menyusun system prompt: %w", err)
	}
	return data, nil
}

// LoadError mengembalikan error pemuatan prompt files (nil bila sukses).
func LoadError() error {
	load()
	return loadErr
}
