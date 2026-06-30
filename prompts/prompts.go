// Package prompts memusatkan seluruh prompt AI yang dipakai backend.
//
// Semua berkas .yml di folder ini di-embed ke dalam binary memakai go:embed,
// sehingga prompt tetap tersentralisasi di satu tempat (mudah diubah) sekaligus
// selalu tersedia di runtime tanpa bergantung pada working directory atau
// menyalin folder ke image Docker.
//
// Dua sumber prompt:
//   - prompts.yml      : kumpulan prompt singkat/berplaceholder (per-section).
//   - ai_prompt.yml    : konfigurasi system prompt utama chat (struktur kaya),
//     diakses lewat SystemPromptYAML().
package prompts

import (
	"embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed *.yml
var files embed.FS

// promptsData memetakan section -> key -> teks prompt.
var (
	promptsData map[string]map[string]string
	loadOnce    sync.Once
	loadErr     error
)

func load() {
	loadOnce.Do(func() {
		data, err := files.ReadFile("prompts.yml")
		if err != nil {
			loadErr = fmt.Errorf("prompts: gagal membaca prompts.yml: %w", err)
			return
		}
		if err := yaml.Unmarshal(data, &promptsData); err != nil {
			loadErr = fmt.Errorf("prompts: gagal parse prompts.yml: %w", err)
			return
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

// SystemPromptYAML mengembalikan isi mentah ai_prompt.yml (system prompt chat
// yang kaya struktur) untuk di-parse oleh pemanggil.
func SystemPromptYAML() ([]byte, error) {
	return files.ReadFile("ai_prompt.yml")
}

// LoadError mengembalikan error pemuatan prompts.yml (nil bila sukses).
func LoadError() error {
	load()
	return loadErr
}
