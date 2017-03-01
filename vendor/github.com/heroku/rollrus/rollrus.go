package rollrus

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/stvp/roll"
)

var defaultTriggerLevels = []log.Level{
	log.ErrorLevel,
	log.FatalLevel,
	log.PanicLevel,
}

// wellKnownErrorFields are fields that are expected to be of type `error`
// in priority order.
var wellKnownErrorFields = []string{
	"err", "error",
}

// Hook wrapper for the rollbar Client
// May be used as a rollbar client itself
type Hook struct {
	roll.Client
	triggers []log.Level
}

// NewHook for use with when adding to you own logger instance. Uses the defualt
// report levels.
func NewHook(token string, env string) *Hook {
	return NewHookForLevels(token, env, defaultTriggerLevels)
}

// NewHookForLevels provided by the caller. Otherwise works like NewHook.
func NewHookForLevels(token string, env string, levels []log.Level) *Hook {
	return &Hook{
		Client:   roll.New(token, env),
		triggers: levels,
	}
}

// SetupLogging for use on Heroku. If token is not an empty string a rollbar
// hook is added with the environment set to env. The log formatter is set to a
// TextFormatter with timestamps disabled.
func SetupLogging(token, env string) {
	setupLogging(token, env, defaultTriggerLevels)
}

// SetupLoggingForLevels works like SetupLogging, but allows you to
// set the levels on which to trigger this hook.
func SetupLoggingForLevels(token, env string, levels []log.Level) {
	setupLogging(token, env, levels)
}

func setupLogging(token, env string, levels []log.Level) {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	if token != "" {
		log.AddHook(NewHookForLevels(token, env, levels))
	}
}

// ReportPanic attempts to report the panic to rollbar using the provided
// client and then re-panic. If it can't report the panic it will print an
// error to stderr.
func (r *Hook) ReportPanic() {
	if p := recover(); p != nil {
		if _, err := r.Client.Critical(fmt.Errorf("panic: %q", p), nil); err != nil {
			fmt.Fprintf(os.Stderr, "reporting_panic=false err=%q\n", err)
		}
		panic(p)
	}
}

// ReportPanic attempts to report the panic to rollbar if the token is set
func ReportPanic(token, env string) {
	if token != "" {
		h := &Hook{Client: roll.New(token, env)}
		h.ReportPanic()
	}
}

// Fire the hook. This is called by Logrus for entries that match the levels
// returned by Levels(). See below.
func (r *Hook) Fire(entry *log.Entry) error {
	cause, trace := extractError(entry)
	m := convertFields(entry.Data)
	if _, exists := m["time"]; !exists {
		m["time"] = entry.Time.Format(time.RFC3339)
	}

	return r.report(entry, cause, m, trace)
}

func (r *Hook) report(entry *log.Entry, cause error, m map[string]string, trace []uintptr) (err error) {
	hasTrace := len(trace) > 0
	level := entry.Level

	switch {
	case hasTrace && level == log.FatalLevel:
		_, err = r.Client.CriticalStack(cause, trace, m)
	case hasTrace && level == log.PanicLevel:
		_, err = r.Client.CriticalStack(cause, trace, m)
	case hasTrace && level == log.ErrorLevel:
		_, err = r.Client.ErrorStack(cause, trace, m)
	case hasTrace && level == log.WarnLevel:
		_, err = r.Client.WarningStack(cause, trace, m)
	case level == log.FatalLevel || level == log.PanicLevel:
		_, err = r.Client.Critical(cause, m)
	case level == log.ErrorLevel:
		_, err = r.Client.Error(cause, m)
	case level == log.WarnLevel:
		_, err = r.Client.Warning(cause, m)
	case level == log.InfoLevel:
		_, err = r.Client.Info(entry.Message, m)
	case level == log.DebugLevel:
		_, err = r.Client.Debug(entry.Message, m)
	}
	return err
}

// Levels returns the logrus log levels that this hook handles
func (r *Hook) Levels() []log.Level {
	if r.triggers == nil {
		return defaultTriggerLevels
	}
	return r.triggers
}

// convertFields converts from log.Fields to map[string]string so that we can
// report extra fields to Rollbar
func convertFields(fields log.Fields) map[string]string {
	m := make(map[string]string)
	for k, v := range fields {
		switch t := v.(type) {
		case time.Time:
			m[k] = t.Format(time.RFC3339)
		default:
			if s, ok := v.(fmt.Stringer); ok {
				m[k] = s.String()
			} else {
				m[k] = fmt.Sprintf("%+v", t)
			}
		}
	}

	return m
}

// extractError attempts to extract an error from a well known field, err or error
func extractError(entry *log.Entry) (error, []uintptr) {
	var trace []uintptr
	fields := entry.Data

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	for _, f := range wellKnownErrorFields {
		e, ok := fields[f]
		if !ok {
			continue
		}
		err, ok := e.(error)
		if !ok {
			continue
		}

		cause := errors.Cause(err)
		tracer, ok := err.(stackTracer)
		if ok {
			return cause, copyStackTrace(tracer.StackTrace())
		}
		return cause, trace
	}

	// when no error found, default to the logged message.
	return fmt.Errorf(entry.Message), trace
}

func copyStackTrace(trace errors.StackTrace) (out []uintptr) {
	for _, frame := range trace {
		out = append(out, uintptr(frame))
	}
	return
}
