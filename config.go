package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// AppConfig stores user editable application settings.
type AppConfig struct {
	Port int `json:"port"`
	Auth bool `json:"auth"`
	Password string `json:"password"`
	AutoOpenBrowser bool `json:"auto_open_browser"`
	MaxUploadMB int64 `json:"max_upload_mb"`
}

func defaultConfig() AppConfig {
	return AppConfig{Port:8000, Auth:true, AutoOpenBrowser:true, MaxUploadMB:512}
}

func configPath() string {
	base, err := os.UserConfigDir()
	if err != nil { base = "." }
	return filepath.Join(base, "LAN Clipboard", "config.json")
}

func loadConfig() AppConfig {
	cfg := defaultConfig()
	b, err := os.ReadFile(configPath())
	if err == nil { _ = json.Unmarshal(b, &cfg) }
	return cfg
}

func saveConfig(cfg AppConfig) error {
	p := configPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil { return err }
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil { return err }
	return os.WriteFile(p, append(b, '\n'), 0644)
}

func appDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		h, _ := os.UserHomeDir(); return filepath.Join(h,"Library","Application Support","LAN Clipboard")
	case "windows":
		d, _ := os.UserConfigDir(); return filepath.Join(d,"LAN Clipboard")
	default:
		h, _ := os.UserHomeDir(); return filepath.Join(h,".local","share","lan-clipboard")
	}
}
