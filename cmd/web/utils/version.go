package utils

import (
	"log/slog"
	"runtime/debug"
)

var Version = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			slog.Info("got debug", "setting", setting)
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "dev"
}()
