package logger

// Entry is a log entry.
// Entry contains all fields required by our structured logging.
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

// WithContext adds context to a log entry.
func (e *Entry) WithContext(context string) *Entry {
	if !isJSON(context) {
		context = "{}"
	}

	e.ContextAsString = context

	return e
}
