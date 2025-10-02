package main

import "log/slog"

// Slogger implements the github.com/mchipperfield/bollocks/pkg/log.Logger interface using the embedded slog Logger.
type Slogger struct {
	*slog.Logger
}

func (s *Slogger) Log(msg string, keyvals ...any) error {
	s.Logger.Info(msg, keyvals...)
	return nil
}
