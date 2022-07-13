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

type Logger struct {
	Entries   []*Entry
	env       string
	request   *http.Request
	startTime time.Time
	traceID   string
	writer    io.Writer
	service   string
}

type Config struct {
	Writer  io.Writer
	Env     string
	Service string
}

type Entry struct {
	Args            string `json:"args"`
	CauserID        string `json:"causer_id"`
	CauserType      string `json:"causer_type"`
	ContextAsString string `json:"context_as_string"`
	DataId          string `json:"data_id"`
	DataType        string `json:"data_type"`
	Datetime        string `json:"datetime"`
	Delta           int    `json:"delta"`
	Env             string `json:"env"`
	Impersonator    string `json:"impersonator"`
	Level           string `json:"level"`
	Message         string `json:"message"`
	ProcessContext  string `json:"process_context"`
	ProcessStart    string `json:"process_start"`
	RemoteAddress   string `json:"remote_address"`
	RequestMethod   string `json:"request_method"`
	RequestQuery    string `json:"request_query"`
	RequestURL      string `json:"request_url"`
	Service         string `json:"service"`
	StatusCode      int    `json:"status_code"`
	TraceID         string `json:"trace_id"`
	Type            string `json:"type"`
}

func (l *Logger) Debug(msg string) *Entry {
	return l.Log(Debug, msg)
}

func (l *Logger) Info(msg string) *Entry {
	return l.Log(Info, msg)
}

func (l *Logger) Warn(msg string) *Entry {
	return l.Log(Warn, msg)
}

func (l *Logger) Error(msg string) *Entry {
	return l.Log(Error, msg)
}

func (l *Logger) Log(lvl level, msg string) *Entry {
	entry := l.buildEntry(lvl, msg)
	l.Entries = append(l.Entries, entry)

	return entry
}

func (e *Entry) WithContext(context string) *Entry {
	if !isJSON(context) {
		context = "{}"
	}

	e.ContextAsString = context

	return e
}

func (l *Logger) WithRequest(request *http.Request) *Logger {
	if request == nil {
		return l
	}

	l.request = request

	if request.Header.Get("x-trace-id") != "" {
		l.traceID = request.Header.Get("x-trace-id")
	}

	return l
}

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

func FromContext(ctx context.Context) (*Logger, error) {
	logger, ok := ctx.Value(contextKeyRequest).(*Logger)
	if !ok {
		return nil, ErrNoRequest
	}

	return logger, nil
}
