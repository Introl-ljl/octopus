package cli

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrUnauthorized   = &CLIError{ExitCode: ExitCodeAuth, Message: "not authenticated. Use 'octopus remote login' first"}
	ErrForbidden      = &CLIError{ExitCode: ExitCodeAuth, Message: "permission denied"}
	ErrNotImplemented = errors.New("not implemented")
)

type ExitCode int

const (
	ExitCodeSuccess       ExitCode = 0
	ExitCodeGeneral       ExitCode = 1
	ExitCodeAuth          ExitCode = 2
	ExitCodeInput         ExitCode = 3
	ExitCodeServerError   ExitCode = 4
	ExitCodeNetworkError  ExitCode = 5
)

type CLIError struct {
	ExitCode ExitCode
	Message  string
	Cause    error
}

func (e *CLIError) Error() string {
	if e.Cause != nil && e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "unknown error"
}

func (e *CLIError) Unwrap() error {
	return e.Cause
}

func NewCLIError(code ExitCode, msg string) *CLIError {
	return &CLIError{ExitCode: code, Message: msg}
}

func WrapCLIError(code ExitCode, msg string, cause error) *CLIError {
	return &CLIError{ExitCode: code, Message: msg, Cause: cause}
}

func GetExitCode(err error) ExitCode {
	if err == nil {
		return ExitCodeSuccess
	}
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr.ExitCode
	}
	return ExitCodeGeneral
}

func ExitWithError(err error) {
	code := GetExitCode(err)
	fmt.Fprintln(os.Stderr, "Error:", err)
	osExit(int(code))
}

var osExit = func(code int) {
	panic(fmt.Sprintf("exit %d", code))
}

func ExitWithCode(code int) {
	osExit(code)
}

func SetExitOverride(fn func(code int)) {
	osExit = fn
}
