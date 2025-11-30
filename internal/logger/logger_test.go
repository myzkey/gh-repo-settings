package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"quiet", LevelQuiet},
		{"normal", LevelNormal},
		{"verbose", LevelVerbose},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.level)
			if l == nil {
				t.Fatal("New() returned nil")
			}
			if l.level != tt.level {
				t.Errorf("level = %d, want %d", l.level, tt.level)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	l := New(LevelNormal)
	l.SetLevel(LevelVerbose)
	if l.level != LevelVerbose {
		t.Errorf("level = %d, want %d", l.level, LevelVerbose)
	}
}

func TestSetOutput(t *testing.T) {
	l := New(LevelNormal)
	buf := &bytes.Buffer{}
	l.SetOutput(buf)

	l.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("output = %q, want to contain %q", buf.String(), "test message")
	}
}

func TestSetErrorOutput(t *testing.T) {
	l := New(LevelNormal)
	buf := &bytes.Buffer{}
	l.SetErrorOutput(buf)

	l.Warn("warning message")

	if !strings.Contains(buf.String(), "warning message") {
		t.Errorf("output = %q, want to contain %q", buf.String(), "warning message")
	}
}

func TestDebug(t *testing.T) {
	t.Run("verbose level shows debug", func(t *testing.T) {
		l := New(LevelVerbose)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Debug("debug message %s", "arg")

		if !strings.Contains(buf.String(), "debug message arg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "debug message arg")
		}
		if !strings.Contains(buf.String(), "[DEBUG]") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "[DEBUG]")
		}
	})

	t.Run("normal level hides debug", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Debug("debug message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})

	t.Run("quiet level hides debug", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Debug("debug message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestInfo(t *testing.T) {
	t.Run("normal level shows info", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Info("info message %d", 42)

		if !strings.Contains(buf.String(), "info message 42") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "info message 42")
		}
	})

	t.Run("quiet level hides info", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Info("info message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})

	t.Run("verbose level shows info", func(t *testing.T) {
		l := New(LevelVerbose)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Info("info message")

		if !strings.Contains(buf.String(), "info message") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "info message")
		}
	})
}

func TestSuccess(t *testing.T) {
	t.Run("normal level shows success", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Success("success message")

		output := buf.String()
		if !strings.Contains(output, "success message") {
			t.Errorf("output = %q, want to contain %q", output, "success message")
		}
	})

	t.Run("quiet level hides success", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Success("success message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestWarn(t *testing.T) {
	t.Run("normal level shows warning", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetErrorOutput(buf)

		l.Warn("warning message")

		output := buf.String()
		if !strings.Contains(output, "warning message") {
			t.Errorf("output = %q, want to contain %q", output, "warning message")
		}
	})

	t.Run("quiet level hides warning", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetErrorOutput(buf)

		l.Warn("warning message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestError(t *testing.T) {
	t.Run("normal level shows error", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetErrorOutput(buf)

		l.Error("error message")

		output := buf.String()
		if !strings.Contains(output, "error message") {
			t.Errorf("output = %q, want to contain %q", output, "error message")
		}
	})

	t.Run("quiet level shows error", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetErrorOutput(buf)

		l.Error("error message")

		output := buf.String()
		if !strings.Contains(output, "error message") {
			t.Errorf("output = %q, want to contain %q", output, "error message")
		}
	})
}

func TestPrint(t *testing.T) {
	t.Run("normal level shows print", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Print("raw message %s", "test")

		if buf.String() != "raw message test" {
			t.Errorf("output = %q, want %q", buf.String(), "raw message test")
		}
	})

	t.Run("quiet level hides print", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Print("raw message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestPrintln(t *testing.T) {
	t.Run("normal level shows println", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Println("message", "with", "args")

		if !strings.Contains(buf.String(), "message with args") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "message with args")
		}
	})

	t.Run("quiet level hides println", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Println("message")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestProgress(t *testing.T) {
	t.Run("normal level shows progress", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Progress("loading... ")

		if buf.String() != "loading... " {
			t.Errorf("output = %q, want %q", buf.String(), "loading... ")
		}
	})

	t.Run("quiet level hides progress", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.Progress("loading... ")

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestProgressDone(t *testing.T) {
	t.Run("normal level shows progress done", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.ProgressDone()

		output := buf.String()
		if output == "" {
			t.Error("output should not be empty")
		}
	})

	t.Run("quiet level hides progress done", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.ProgressDone()

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestProgressFail(t *testing.T) {
	t.Run("normal level shows progress fail", func(t *testing.T) {
		l := New(LevelNormal)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.ProgressFail()

		output := buf.String()
		if output == "" {
			t.Error("output should not be empty")
		}
	})

	t.Run("quiet level hides progress fail", func(t *testing.T) {
		l := New(LevelQuiet)
		buf := &bytes.Buffer{}
		l.SetOutput(buf)

		l.ProgressFail()

		if buf.String() != "" {
			t.Errorf("output = %q, want empty", buf.String())
		}
	})
}

func TestGlobalFunctions(t *testing.T) {
	// Save original default logger
	origLogger := defaultLogger
	defer func() { defaultLogger = origLogger }()

	t.Run("SetDefaultLevel", func(t *testing.T) {
		SetDefaultLevel(LevelVerbose)
		if defaultLogger.level != LevelVerbose {
			t.Errorf("defaultLogger.level = %d, want %d", defaultLogger.level, LevelVerbose)
		}
	})

	t.Run("Default", func(t *testing.T) {
		l := Default()
		if l != defaultLogger {
			t.Error("Default() should return defaultLogger")
		}
	})

	t.Run("global Debug", func(t *testing.T) {
		buf := &bytes.Buffer{}
		defaultLogger = New(LevelVerbose)
		defaultLogger.SetOutput(buf)

		Debug("global debug %s", "msg")

		if !strings.Contains(buf.String(), "global debug msg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "global debug msg")
		}
	})

	t.Run("global Info", func(t *testing.T) {
		buf := &bytes.Buffer{}
		defaultLogger = New(LevelNormal)
		defaultLogger.SetOutput(buf)

		Info("global info %s", "msg")

		if !strings.Contains(buf.String(), "global info msg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "global info msg")
		}
	})

	t.Run("global Success", func(t *testing.T) {
		buf := &bytes.Buffer{}
		defaultLogger = New(LevelNormal)
		defaultLogger.SetOutput(buf)

		Success("global success %s", "msg")

		if !strings.Contains(buf.String(), "global success msg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "global success msg")
		}
	})

	t.Run("global Warn", func(t *testing.T) {
		buf := &bytes.Buffer{}
		defaultLogger = New(LevelNormal)
		defaultLogger.SetErrorOutput(buf)

		Warn("global warn %s", "msg")

		if !strings.Contains(buf.String(), "global warn msg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "global warn msg")
		}
	})

	t.Run("global Error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		defaultLogger = New(LevelNormal)
		defaultLogger.SetErrorOutput(buf)

		Error("global error %s", "msg")

		if !strings.Contains(buf.String(), "global error msg") {
			t.Errorf("output = %q, want to contain %q", buf.String(), "global error msg")
		}
	})
}

func TestLevelConstants(t *testing.T) {
	// Ensure level constants have expected values
	if LevelQuiet != 0 {
		t.Errorf("LevelQuiet = %d, want 0", LevelQuiet)
	}
	if LevelNormal != 1 {
		t.Errorf("LevelNormal = %d, want 1", LevelNormal)
	}
	if LevelVerbose != 2 {
		t.Errorf("LevelVerbose = %d, want 2", LevelVerbose)
	}
}
