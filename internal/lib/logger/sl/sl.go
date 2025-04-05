package sl

import "log/slog"

// Err makes slog.Attr for error
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}
