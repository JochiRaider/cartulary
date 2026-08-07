package config

func validate(cfg document) (document, error) {
	normalized := cfg
	diagnostics := validateConfigStructure(&normalized, configPresence{}, nil)
	if len(diagnostics) > 0 {
		return document{}, newDiagnosticsError(diagnostics)
	}
	return normalized, nil
}

func validateForStartup(cfg document) (document, error) {
	diagnostics := validateStartupFilesystemRoots(cfg.Roots)
	if len(diagnostics) > 0 {
		return document{}, newDiagnosticsError(diagnostics)
	}
	return cfg, nil
}
