package errors

import (
	"context"
	"fmt"
	"log"
)

// AppError defines the structure for an application-specific error.
// It includes a stack of function names, an error message, and an optional cause.
type AppError struct {
	FuncStack []string
	Message   string
	Cause     error
}

// Error implements the error interface for the AppError struct.
// It formats the error message including the function stack and the cause if available.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.FuncStack, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.FuncStack, e.Message)
}

// UnwrapError recursively unwraps errors that implement the Unwrap method.
// It returns the innermost error in the chain.
func UnwrapError(err error) error {
	for {
		if unwrappedErr, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrappedErr.Unwrap()
		} else {
			return err
		}
	}
}

// New creates a new AppError with the provided function name, message, and underlying cause.
func New(funcName, message string, cause error) error {
	return &AppError{
		FuncStack: []string{funcName},
		Message:   message,
		Cause:     cause,
	}
}

// Wrap takes an existing error and adds the current function name to its stack trace.
func Wrap(funcName string, err error) error {
	if appErr, ok := err.(*AppError); ok {
		appErr.FuncStack = append(appErr.FuncStack, funcName)
		return appErr
	}
	return &AppError{
		FuncStack: []string{funcName},
		Message:   err.Error(),
		Cause:     err,
	}
}

// HandleDeadlineExceededError checks if the given error is a context deadline exceeded error.
func HandleDeadlineExceededError(packageName string, err error) error {
	funcName := packageName + "HandleDeadlineExceededError"
	if err == context.DeadlineExceeded {
		log.Printf("Operation timed out: %v", err)
		return New(funcName, "Operation timed out: ", err)
	}
	return nil
}
