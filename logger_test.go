package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	for _, line := range bytes.Split(writer.Bytes(), []byte{'\n'}) {
		var entry Entry

		if !isJSON(string(line)) {
			continue
		}

		err := json.Unmarshal(line, &entry)
		is.NoErr(err)

		entries = append(entries, entry)
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

func TestLogger_DecorateEntries(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	entries := []*Entry{{
		StatusCode: 0,
		Service:    "_",
		Env:        "prod",
	}}

	lgr := Logger{Entries: entries}

	lgr.DecorateEntries(func(entry *Entry) *Entry {
		entry.StatusCode = 404
		entry.Service = "my-service"

		return entry
	})

	is.Equal(404, entries[0].StatusCode)
	is.Equal("my-service", entries[0].Service)
	is.Equal("prod", entries[0].Env)
}

func Test_isJSON(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "works on basic json",
			args: args{
				str: "{}",
			},
			want: true,
		},
		{
			name: "works on complicated json",
			args: args{
				str: `
            [
              { "cat": 3, "dog": true },
              {
                "nested": {
                  "mouse": "squeak",
                  "some-array": ["hi", "bye"]
                }
              }
            ]
				    `,
			},
			want: true,
		},
		{
			name: "empty string is not json",
			args: args{
				str: "",
			},
			want: false,
		},
		{
			name: "invalid json is not json",
			args: args{
				str: "{whoops}",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isJSON(tt.args.str); got != tt.want {
				t.Errorf("isJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	writer := &bytes.Buffer{}
	cfg := Config{
		Writer:  writer,
		Env:     "some-env",
		Service: "some-service",
	}

	middleware := Middleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr, err := FromContext(r.Context())
		is.NoErr(err)
		lgr.Info("some message")
		lgr.Warn("some warning").WithContext(`{"warning": "oh no"}`)
	})

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://my.app/some-path", nil)
	req.Header.Add("x-trace-id", "my-trace")

	middleware(handler).ServeHTTP(resp, req)

	entries := []Entry{}

	for _, line := range bytes.Split(writer.Bytes(), []byte{'\n'}) {
		var entry Entry

		if !isJSON(string(line)) {
			continue
		}

		err := json.Unmarshal(line, &entry)
		is.NoErr(err)

		entries = append(entries, entry)
	}

	firstLog := entries[0]
	secondLog := entries[1]

	is.Equal("some message", firstLog.Message)
	is.Equal(Info.String(), firstLog.Level)
	is.Equal("", firstLog.ContextAsString)
	is.Equal("my-trace", firstLog.TraceID)
	is.Equal("my.app/some-path", firstLog.RequestURL)

	is.Equal("some warning", secondLog.Message)
	is.Equal(Warn.String(), secondLog.Level)
	is.Equal(`{"warning": "oh no"}`, secondLog.ContextAsString)
	is.Equal("my-trace", secondLog.TraceID)
	is.Equal("my.app/some-path", secondLog.RequestURL)

	// clear buffer
	writer.Reset()

	// ensure middleware gives a new logger to next request

	req = httptest.NewRequest(http.MethodGet, "https://my.app/some-other-path", nil)
	req.Header.Add("x-trace-id", "my-new-trace")

	middleware(handler).ServeHTTP(resp, req)

	entries = []Entry{}

	for _, line := range bytes.Split(writer.Bytes(), []byte{'\n'}) {
		var entry Entry

		if !isJSON(string(line)) {
			continue
		}

		err := json.Unmarshal(line, &entry)
		is.NoErr(err)

		entries = append(entries, entry)
	}

	firstLog = entries[0]
	secondLog = entries[1]

	is.Equal("some message", firstLog.Message)
	is.Equal(Info.String(), firstLog.Level)
	is.Equal("", firstLog.ContextAsString)
	is.Equal("my-new-trace", firstLog.TraceID)
	is.Equal("my.app/some-other-path", firstLog.RequestURL)

	is.Equal("some warning", secondLog.Message)
	is.Equal(Warn.String(), secondLog.Level)
	is.Equal(`{"warning": "oh no"}`, secondLog.ContextAsString)
	is.Equal("my-new-trace", secondLog.TraceID)
	is.Equal("my.app/some-other-path", secondLog.RequestURL)
}
