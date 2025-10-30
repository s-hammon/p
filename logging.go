package p

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

func Colorize(color int, text string) string {
	return Format("\x1b[%dm%v\x1b[0m", color, text)
}

// Logging specific to GCP
type Severity int32

const (
	Debug Severity = iota + 1
	Info
	Notice
	Warning
	Error
	Critical
	Alert
	Emergency
)

func (s Severity) MarshalJSON() ([]byte, error) {
	switch s {
	default:
		return []byte(`"UNKNOWN"`), fmt.Errorf("unknown severity: %d", s)
	case Debug:
		return []byte(`"DEBUG"`), nil
	case Info:
		return []byte(`"INFO"`), nil
	case Notice:
		return []byte(`"NOTICE"`), nil
	case Warning:
		return []byte(`"WARNING"`), nil
	case Error:
		return []byte(`"ERROR"`), nil
	case Critical:
		return []byte(`"CRITICAL"`), nil
	case Alert:
		return []byte(`"ALERT"`), nil
	case Emergency:
		return []byte(`"EMERGENCY"`), nil
	}
}

type Entry struct {
	Message  string          `json:"message"`
	Severity Severity        `json:"severity,omitempty"`
	Trace    json.RawMessage `json:"logging.googleapis.com/trace,omitempty"`
}

type GcpLogger struct {
	out, err io.Writer
	mu       sync.Mutex
	trace    json.RawMessage
}

func (l *GcpLogger) Print(v ...any) {
	logGCP(Info, l, fmt.Sprint(v...))
}

func (l *GcpLogger) Debug(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Debug, l, s)
}

func (l *GcpLogger) Info(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Info, l, s)
}

func (l *GcpLogger) Notice(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Notice, l, s)
}

func (l *GcpLogger) Warning(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Warning, l, s)
}

func (l *GcpLogger) Error(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Error, l, s)
}

func (l *GcpLogger) Critical(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Critical, l, s)
}

func (l *GcpLogger) Alert(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Alert, l, s)
}

func (l *GcpLogger) Emergency(s string, v ...any) {
	if len(v) > 0 {
		s = Format(s, v...)
	}
	logGCP(Emergency, l, s)
}

func (l *GcpLogger) writer(s Severity) io.Writer {
	if s >= Error {
		if l.err != nil {
			return l.err
		}
		return os.Stderr
	}
	if l.out != nil {
		return l.out
	}
	return os.Stdout
}

func logGCP(s Severity, l *GcpLogger, msg string) string {
	entry := Entry{msg, s, l.trace}
	fmt.Println(entry)
	enc := json.NewEncoder(l.writer(s))
	enc.SetEscapeHTML(false)
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = enc.Encode(entry)
	return msg
}
