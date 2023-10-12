package commons

import (
	"io"
	"log/slog"
	"os"
)

func GetStructuredLogger(minLevel slog.Level, out io.Writer) *slog.Logger {
	levelVar := new(slog.LevelVar)
	levelVar.Set(slog.LevelDebug)
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     levelVar,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
				if a.Value.String() == slog.LevelWarn.String() {
					a.Value = slog.StringValue("WARNING")
				}
			}
			return a
		},
	}))
}
