package main

import "net/http"

// getConfig returns current user configuration.
func (a *App) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, loadConfig())
}

// updateConfig stores user configuration.
// Runtime reload hooks are added in the next step when the server state is
// connected with the persisted configuration.
func (a *App) updateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg AppConfig
	if err := decodeJSON(r, &cfg, 8192); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if cfg.Port <= 0 {
		cfg.Port = 8000
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = 512
	}
	if err := saveConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "config": cfg})
}
