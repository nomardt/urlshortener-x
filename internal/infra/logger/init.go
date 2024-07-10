package logger

import "go.uber.org/zap"

// Creating a global Log variable
var Log *zap.Logger = zap.NewNop()

// Initialize the global variable Log
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
