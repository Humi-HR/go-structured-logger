package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// level is the log level.
type level int64

const (
	Debug level = iota
	Info
	Warn
	Error
)

// contextKey is used to register the logger into context.
type contextKey string

var (
	// required global. not exported.
	// nolint
	contextKeyRequest     = contextKey("request")
	ErrNoRequest          = errors.New("no request")
	ErrInvalidJSONContext = errors.New("invalid JSON context")
)

func (s level) String() string {
	switch s {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	}

	return "unknown"
}

// Logger is main entity in this package.
// Logger handles writing of logs.
// In an HTTP context, a new logger is created per request.
type Logger struct {
	Entries   []*Entry
	env       string
	request   *http.Request
	startTime time.Time
	traceID   string
	writer    io.Writer
	service   string
}

// Config is used to configure the logger.
// All log entries for the logger will use this configuration.
type Config struct {
	Writer  io.Writer
	Env     string
	Service string
}

// Debug messages are used to debug the application.
// They will not appear in production.
func (l *Logger) Debug(msg string) *Entry {
	return l.Log(Debug, msg)
}

// Info messages are standard log messages that give information about
// what is happening in the application.
func (l *Logger) Info(msg string) *Entry {
	return l.Log(Info, msg)
}

// Warn messages are used to warn that something is happening that we'd rather
// not happen. It's not an error, but it's a cause for some concern.
// Ex: a deprecated method is used.
func (l *Logger) Warn(msg string) *Entry {
	return l.Log(Warn, msg)
}

// Error messages tell us that something went wrong.
func (l *Logger) Error(msg string) *Entry {
	return l.Log(Error, msg)
}

func (l *Logger) Log(lvl level, msg string) *Entry {
	entry := l.buildEntry(lvl, msg)
	l.Entries = append(l.Entries, entry)

	return entry
}

// WithRequest adds a request to the logger.
// It also sets the trace ID if one exists.
func (l *Logger) WithRequest(request *http.Request) *Logger {
	l.request = request

	if request != nil && request.Header.Get("x-trace-id") != "" {
		l.traceID = request.Header.Get("x-trace-id")
	}

	return l
}

// buildEntry creates an Entry with all possible values.
func (l *Logger) buildEntry(lvl level, msg string) *Entry {
	startTime := l.startTime.Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)
	delta := time.Since(l.startTime)

	remoteAddress := ""
	requestMethod := ""
	requestQuery := ""
	requestURL := ""

	if l.request != nil {
		remoteAddress = l.request.RemoteAddr
		requestMethod = l.request.Method
		requestQuery = l.request.URL.RawQuery
		requestURL = l.request.Host + l.request.URL.Path
	}

	return &Entry{
		Args:           strings.Join(os.Args, " "),
		Datetime:       now,
		Delta:          int(delta.Milliseconds()),
		Env:            l.env,
		Level:          lvl.String(),
		Message:        msg,
		ProcessContext: "request",
		ProcessStart:   startTime,
		RemoteAddress:  remoteAddress,
		RequestMethod:  requestMethod,
		RequestQuery:   requestQuery,
		RequestURL:     requestURL,
		Service:        l.service,
		TraceID:        l.traceID,
		Type:           "general",
	}
}

// DecorateEntries is used to modify existing entries.
// It should be called after all log entries are created because it does not apply to future entries.
func (l *Logger) DecorateEntries(decorators ...func(*Entry) *Entry) {
	entries := l.Entries
	for i := range l.Entries {
		for _, decorator := range decorators {
			entries[i] = decorator(entries[i])
		}
	}

	l.Entries = entries
}

// Flush writes all buffered log entries.
// The buffer is then flushed.
func (l *Logger) Flush() {
	if l.writer == nil {
		l.Entries = []*Entry{}
		return
	}

	for _, e := range l.Entries {
		data, err := json.Marshal(e)
		if err == nil {
			fmt.Fprintln(l.writer, string(data))
		}
	}

	l.Entries = []*Entry{}
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// NewLogger creates a new logger for use in an application.
// If you are logging in and HTTP context, use Middleware instead.
func NewLogger(cfg Config) *Logger {
	traceID := uuid.New().String()

	return &Logger{
		Entries:   []*Entry{},
		env:       cfg.Env,
		service:   cfg.Service,
		startTime: time.Now(),
		traceID:   traceID,
		writer:    cfg.Writer,
	}
}

// Middleware creates a middleware for use in an HTTP context.
// Each request will get its own logger.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := NewLogger(cfg).WithRequest(r)
			defer logger.Flush()

			ctx := context.WithValue(r.Context(), contextKeyRequest, logger)

			// wrap the response writer so we can read its values after the request completes
			wrappedResponseWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrappedResponseWriter, r.WithContext(ctx))
			logger.DecorateEntries(func(entry *Entry) *Entry {
				entry.StatusCode = wrappedResponseWriter.Status()
				return entry
			})
		})
	}
}

// FromContext retrieves a logger from a context object.
// This is how we use the logger in our HTTP handlers.
func FromContext(ctx context.Context) (*Logger, error) {
	logger, ok := ctx.Value(contextKeyRequest).(*Logger)
	if !ok {
		return nil, ErrNoRequest
	}

	return logger, nil
}
