package main

import "flag"

// RuntimeOptions combines persisted settings with optional command line overrides.
type RuntimeOptions struct {
	Port int
	DataDir string
	Auth bool
	Password string
	Open bool
	MaxUploadMB int64
}

// loadRuntimeOptions loads config.json first, then applies explicit CLI flags.
func loadRuntimeOptions() RuntimeOptions {
	cfg := loadConfig()
	var opt RuntimeOptions
	port := flag.Int("port", cfg.Port, "HTTP listen port")
	dataDir := flag.String("data-dir", appDataDir(), "application data directory")
	auth := flag.Bool("auth", cfg.Auth, "require access password")
	password := flag.String("password", cfg.Password, "access password")
	open := flag.Bool("open", cfg.AutoOpenBrowser, "open browser")
	maxUpload := flag.Int64("max-upload", cfg.MaxUploadMB, "maximum upload size MB")
	flag.Parse()
	opt.Port = *port
	opt.DataDir = *dataDir
	opt.Auth = *auth
	opt.Password = *password
	opt.Open = *open
	opt.MaxUploadMB = *maxUpload
	return opt
}
