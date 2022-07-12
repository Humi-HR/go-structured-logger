package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/matryer/is"
)

func TestLogger_Log(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	writer := &bytes.Buffer{}
	lgr := NewLogger(Config{
		Writer:  writer,
		Env:     "some-env",
		Service: "some-service",
	})

	lgr.Info("some message")
	lgr.Warn("some warning").WithContext(`{"warning": "oh no"}`)
	lgr.Flush()

	entries := []Entry{}

	err := json.Unmarshal(writer.Bytes(), &entries)
	if err != nil {
		t.Fatal(err)
	}

	firstLog := entries[0]
	secondLog := entries[1]

	is.Equal("some message", firstLog.Message)
	is.Equal(Info.String(), firstLog.Level)
	is.Equal("", firstLog.ContextAsString)

	is.Equal("some warning", secondLog.Message)
	is.Equal(Warn.String(), secondLog.Level)
	is.Equal(`{"warning": "oh no"}`, secondLog.ContextAsString)

	is.Equal(firstLog.TraceID, secondLog.TraceID) // Logger should always have the same trace id
}

func TestLogger_WithRequest(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	writer := &bytes.Buffer{}
	lgr := NewLogger(Config{
		Writer:  writer,
		Env:     "some-env",
		Service: "some-service",
	})

	is.Equal(nil, lgr.request)

	req := &http.Request{}
	lgr.WithRequest(req)

	is.Equal(req, lgr.request)
}

func TestLogger_buildEntry(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	writer := &bytes.Buffer{}
	lgr := NewLogger(Config{
		Writer:  writer,
		Env:     "some-env",
		Service: "some-service",
	})

	entry := lgr.buildEntry(Info, "test message")

	is.Equal("test message", entry.Message)
	// start here
	// review build entry and see what to test
	// do not test request stuff
	// try to test that in the middleware... maybe?
}

// func TestLogger_DecorateEntries(t *testing.T) {
// 	type fields struct {
// 		entries   []*Entry
// 		env       string
// 		request   *http.Request
// 		startTime time.Time
// 		traceID   string
// 		writer    io.Writer
// 		service   string
// 	}
// 	type args struct {
// 		decorators []func(*Entry) *Entry
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			l := &Logger{
// 				entries:   tt.fields.entries,
// 				env:       tt.fields.env,
// 				request:   tt.fields.request,
// 				startTime: tt.fields.startTime,
// 				traceID:   tt.fields.traceID,
// 				writer:    tt.fields.writer,
// 				service:   tt.fields.service,
// 			}
// 			l.DecorateEntries(tt.args.decorators...)
// 		})
// 	}
// }

// func TestLogger_Flush(t *testing.T) {
// 	type fields struct {
// 		entries   []*Entry
// 		env       string
// 		request   *http.Request
// 		startTime time.Time
// 		traceID   string
// 		writer    io.Writer
// 		service   string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			l := &Logger{
// 				entries:   tt.fields.entries,
// 				env:       tt.fields.env,
// 				request:   tt.fields.request,
// 				startTime: tt.fields.startTime,
// 				traceID:   tt.fields.traceID,
// 				writer:    tt.fields.writer,
// 				service:   tt.fields.service,
// 			}
// 			l.Flush()
// 		})
// 	}
// }

// func Test_isJSON(t *testing.T) {
// 	type args struct {
// 		str string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			if got := isJSON(tt.args.str); got != tt.want {
// 				t.Errorf("isJSON() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestMiddleware(t *testing.T) {
// 	type args struct {
// 		cfg Config
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func(http.Handler) http.Handler
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			if got := Middleware(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Middleware() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
