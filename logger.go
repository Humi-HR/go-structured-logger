package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type level int64

const (
	Debug level = iota
	Info
	Warn
	Error
)

type contextKey string

var (
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
	entries   []LogEntry
	env       string
	request   *http.Request
	startTime time.Time
	traceID   string
	writer    io.Writer
	service   string
}

type LoggerConfig struct {
	Writer  io.Writer
	Env     string
	Service string
}

type LogEntry struct {
	Args            string `json:"args"`
	CauserId        string `json:"causer_id"`
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

func (l *Logger) Debug(msg string, context string) error {
	return l.Log(Debug, msg, context)
}

func (l *Logger) Info(msg string, context string) error {
	return l.Log(Info, msg, context)
}

func (l *Logger) Warn(msg string, context string) error {
	return l.Log(Warn, msg, context)
}

func (l *Logger) Error(msg string, context string) error {
	return l.Log(Error, msg, context)
}

func (l *Logger) Log(lvl level, msg string, context string) error {
	if context == "" {
		context = "{}"
	}

	if !isJSON(context) {
		reportInvalidContext(context)
		context = "{}"
	}

	entry := l.buildEntry(lvl, msg, context)
	l.entries = append(l.entries, entry)

	return nil
}

func (l *Logger) buildEntry(lvl level, msg string, context string) LogEntry {
	startTime := l.startTime.Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)
	delta := time.Now().Sub(l.startTime)

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

	return LogEntry{
		Args:            strings.Join(os.Args, " "),
		ContextAsString: context,
		Datetime:        now,
		Delta:           int(delta.Milliseconds()),
		Env:             l.env,
		Level:           lvl.String(),
		Message:         msg,
		ProcessContext:  "request",
		ProcessStart:    startTime,
		RemoteAddress:   remoteAddress,
		RequestMethod:   requestMethod,
		RequestQuery:    requestQuery,
		RequestURL:      requestURL,
		Service:         l.service,
		TraceID:         l.traceID,
		Type:            "general",
	}
}

// DecorateEntries is used to modify existing entries.
// It should be called after all log entries are created because it does not apply to future entries.
func (l *Logger) DecorateEntries(decorators ...func(LogEntry) LogEntry) {
	entries := l.entries
	for i := range l.entries {
		for _, decorator := range decorators {
			entries[i] = decorator(entries[i])
		}
	}

	l.entries = entries
}

// Flush writes all buffered log entries.
func (l *Logger) Flush() {
	if l.writer == nil {
		return
	}

	data, err := json.Marshal(l.entries)
	if err != nil {
		log.Println("json failed", err)
	}

	fmt.Fprintln(l.writer, string(data))
	l.entries = []LogEntry{}
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func NewLogger(cfg LoggerConfig, request *http.Request) *Logger {
	traceID := uuid.New().String()
	if request != nil && request.Header.Get("x-trace-id") != "" {
		traceID = request.Header.Get("x-trace-id")
	}

	return &Logger{
		entries:   []LogEntry{},
		env:       cfg.Env,
		request:   request,
		service:   cfg.Service,
		startTime: time.Now(),
		traceID:   traceID,
		writer:    cfg.Writer,
	}
}

func Middleware(cfg LoggerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := NewLogger(cfg, r)
			defer logger.Flush()

			ctx := context.WithValue(r.Context(), contextKeyRequest, logger)

			// wrap the response writer so we can read its values after the request completes
			wrappedResponseWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrappedResponseWriter, r.WithContext(ctx))
			logger.DecorateEntries(func(entry LogEntry) LogEntry {
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

func ReportLoggerNotFound() {
	fmt.Fprintln(os.Stdout, "{\"error\":\"structured logger not found.\"}")
}

func reportInvalidContext(invalidContext string) {
	fmt.Fprintf(
		os.Stdout,
		"{\"error\":\"invalid JSON context.\",\"context\":\"%s\"}\n",
		invalidContext,
	)
}
