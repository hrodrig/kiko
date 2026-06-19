package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, Trace)

	l.Trace("trace %d", 1)
	l.Debug("debug %d", 2)
	l.Info("info %d", 3)
	l.Warn("warn %d", 4)
	l.Error("error %d", 5)

	output := buf.String()
	for _, want := range []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %s", want)
		}
	}
}

func TestLoggerLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, Warn) // only Warn and above

	l.Debug("should not appear")
	l.Info("should not appear")
	l.Warn("this should appear")
	l.Error("this too")

	output := buf.String()
	if strings.Contains(output, "DEBUG") {
		t.Error("DEBUG appeared but level is Warn")
	}
	if strings.Contains(output, "INFO") {
		t.Error("INFO appeared but level is Warn")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("WARN missing")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("ERROR missing")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, Off)
	l.Info("nope")
	if buf.Len() > 0 {
		t.Error("expected no output at Off level")
	}
	l.SetLevel(Debug)
	l.Debug("yes")
	if !strings.Contains(buf.String(), "DEBUG") {
		t.Error("expected DEBUG after SetLevel")
	}
}

func TestLoggerLevel(t *testing.T) {
	l := New(nil, Warn)
	if l.Level() != Warn {
		t.Errorf("Level() = %v; want Warn", l.Level())
	}
}

func TestDefaultLogger(t *testing.T) {
	if Default == nil {
		t.Fatal("Default logger is nil")
	}
	// convenience wrappers should not panic
	Tracef("trace")
	Debugf("debug")
	Infof("info")
	Warnf("warn")
	Errorf("error")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in      string
		want    Level
		wantErr bool
	}{
		{"trace", Trace, false},
		{"debug", Debug, false},
		{"info", Info, false},
		{"warn", Warn, false},
		{"warning", Warn, false},
		{"error", Error, false},
		{"fatal", Fatal, false},
		{"critical", Fatal, false},
		{"off", Off, false},
		{"none", Off, false},
		{"bogus", Info, true},
	}
	for _, tt := range tests {
		got, err := ParseLevel(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseLevel(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestLevelOrder(t *testing.T) {
	if Trace >= Debug {
		t.Error("Trace should be less than Debug")
	}
	if Debug >= Info {
		t.Error("Debug should be less than Info")
	}
	if Info >= Warn {
		t.Error("Info should be less than Warn")
	}
	if Warn >= Error {
		t.Error("Warn should be less than Error")
	}
	if Error >= Fatal {
		t.Error("Error should be less than Fatal")
	}
	if Fatal >= Off {
		t.Error("Fatal should be less than Off")
	}
}
