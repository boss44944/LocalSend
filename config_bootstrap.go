package main

// ensureConfig creates the initial user configuration file when the application
// is started for the first time. Existing user settings are never overwritten.
func ensureConfig() error {
	cfgPath := configPath()
	if _, err := os.Stat(cfgPath); err == nil {
		return nil
	}
	return saveConfig(defaultConfig())
}
