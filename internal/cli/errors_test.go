package cli

import (
	"errors"
	"fmt"
	"testing"
)

func TestCLIErrorCreation(t *testing.T) {
	err := NewCLIError(ExitCodeInput, "invalid input")
	if err.ExitCode != ExitCodeInput {
		t.Errorf("expected ExitCodeInput, got %d", err.ExitCode)
	}
	if err.Message != "invalid input" {
		t.Errorf("expected 'invalid input', got %q", err.Message)
	}
	if err.Error() != "invalid input" {
		t.Errorf("expected 'invalid input', got %q", err.Error())
	}
}

func TestCLIErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("underlying issue")
	err := WrapCLIError(ExitCodeServerError, "server error", cause)
	if err.ExitCode != ExitCodeServerError {
		t.Errorf("expected ExitCodeServerError, got %d", err.ExitCode)
	}
	if !errors.Is(err, cause) {
		t.Error("expected errors.Is to match cause")
	}
	expected := "server error: underlying issue"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestCLIErrorMessageOnly(t *testing.T) {
	err := WrapCLIError(ExitCodeGeneral, "msg", nil)
	if err.Error() != "msg" {
		t.Errorf("expected 'msg', got %q", err.Error())
	}
}

func TestCLIErrorCauseOnly(t *testing.T) {
	err := WrapCLIError(ExitCodeGeneral, "", fmt.Errorf("cause"))
	if err.Error() != "cause" {
		t.Errorf("expected 'cause', got %q", err.Error())
	}
}

func TestCLIErrorEmpty(t *testing.T) {
	err := &CLIError{}
	if err.Error() != "unknown error" {
		t.Errorf("expected 'unknown error', got %q", err.Error())
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		err      error
		expected ExitCode
		name     string
	}{
		{nil, ExitCodeSuccess, "nil error"},
		{NewCLIError(ExitCodeAuth, "auth"), ExitCodeAuth, "auth error"},
		{NewCLIError(ExitCodeInput, "input"), ExitCodeInput, "input error"},
		{NewCLIError(ExitCodeServerError, "server"), ExitCodeServerError, "server error"},
		{NewCLIError(ExitCodeNetworkError, "network"), ExitCodeNetworkError, "network error"},
		{fmt.Errorf("generic"), ExitCodeGeneral, "generic error"},
		{ErrUnauthorized, ExitCodeAuth, "unauthorized sentinel"},
		{ErrForbidden, ExitCodeAuth, "forbidden sentinel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetExitCode(tt.err)
			if code != tt.expected {
				t.Errorf("expected exit code %d, got %d", tt.expected, code)
			}
		})
	}
}

func TestExitWithError(t *testing.T) {
	exitCalled := false
	SetExitOverride(func(code int) {
		exitCalled = true
		if code != int(ExitCodeAuth) {
			t.Errorf("expected exit code %d, got %d", ExitCodeAuth, code)
		}
	})

	ExitWithError(ErrUnauthorized)
	if !exitCalled {
		t.Error("expected exit override to be called")
	}

	// Reset
	SetExitOverride(func(code int) {
		panic(fmt.Sprintf("exit %d", code))
	})
}

func TestExitWithCode(t *testing.T) {
	exitCode := 0
	SetExitOverride(func(code int) {
		exitCode = code
	})

	ExitWithCode(42)
	if exitCode != 42 {
		t.Errorf("expected exit code 42, got %d", exitCode)
	}

	SetExitOverride(func(code int) {
		panic(fmt.Sprintf("exit %d", code))
	})
}

func TestSentinelErrors(t *testing.T) {
	// Verify sentinel errors are *CLIError type
	var cliErr *CLIError
	if !errors.As(ErrUnauthorized, &cliErr) {
		t.Error("ErrUnauthorized should be *CLIError")
	}
	if cliErr.ExitCode != ExitCodeAuth {
		t.Errorf("expected ExitCodeAuth, got %d", cliErr.ExitCode)
	}

	if !errors.As(ErrForbidden, &cliErr) {
		t.Error("ErrForbidden should be *CLIError")
	}
	if cliErr.ExitCode != ExitCodeAuth {
		t.Errorf("expected ExitCodeAuth, got %d", cliErr.ExitCode)
	}
}

func TestNotImplementedError(t *testing.T) {
	if ErrNotImplemented.Error() != "not implemented" {
		t.Errorf("expected 'not implemented', got %q", ErrNotImplemented.Error())
	}
}
