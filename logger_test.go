package logger

import (
	"bytes"
	"encoding/json"
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

	is.Equal("some warning", secondLog.Message)
	is.Equal(Warn.String(), secondLog.Level)

	is.Equal(firstLog.TraceID, secondLog.TraceID) // Logger should always have the same trace id
}

// func TestEntry_WithContext(t *testing.T) {
// 	type fields struct {
// 		Args            string
// 		CauserID        string
// 		CauserType      string
// 		ContextAsString string
// 		DataId          string
// 		DataType        string
// 		Datetime        string
// 		Delta           int
// 		Env             string
// 		Impersonator    string
// 		Level           string
// 		Message         string
// 		ProcessContext  string
// 		ProcessStart    string
// 		RemoteAddress   string
// 		RequestMethod   string
// 		RequestQuery    string
// 		RequestURL      string
// 		Service         string
// 		StatusCode      int
// 		TraceID         string
// 		Type            string
// 	}
// 	type args struct {
// 		context string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   *Entry
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			e := &Entry{
// 				Args:            tt.fields.Args,
// 				CauserID:        tt.fields.CauserID,
// 				CauserType:      tt.fields.CauserType,
// 				ContextAsString: tt.fields.ContextAsString,
// 				DataId:          tt.fields.DataId,
// 				DataType:        tt.fields.DataType,
// 				Datetime:        tt.fields.Datetime,
// 				Delta:           tt.fields.Delta,
// 				Env:             tt.fields.Env,
// 				Impersonator:    tt.fields.Impersonator,
// 				Level:           tt.fields.Level,
// 				Message:         tt.fields.Message,
// 				ProcessContext:  tt.fields.ProcessContext,
// 				ProcessStart:    tt.fields.ProcessStart,
// 				RemoteAddress:   tt.fields.RemoteAddress,
// 				RequestMethod:   tt.fields.RequestMethod,
// 				RequestQuery:    tt.fields.RequestQuery,
// 				RequestURL:      tt.fields.RequestURL,
// 				Service:         tt.fields.Service,
// 				StatusCode:      tt.fields.StatusCode,
// 				TraceID:         tt.fields.TraceID,
// 				Type:            tt.fields.Type,
// 			}
// 			if got := e.WithContext(tt.args.context); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Entry.WithContext() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestLogger_WithRequest(t *testing.T) {
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
// 		request *http.Request
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   *Logger
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
// 			if got := l.WithRequest(tt.args.request); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Logger.WithRequest() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestLogger_buildEntry(t *testing.T) {
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
// 		lvl level
// 		msg string
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 		want   *Entry
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
// 			if got := l.buildEntry(tt.args.lvl, tt.args.msg); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Logger.buildEntry() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

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

// func TestNewLogger(t *testing.T) {
// 	type args struct {
// 		cfg Config
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *Logger
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			if got := NewLogger(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewLogger() = %v, want %v", got, tt.want)
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

// func TestFromContext(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *Logger
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			got, err := FromContext(tt.args.ctx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FromContext() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FromContext() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
