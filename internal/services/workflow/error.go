package workflow

import (
	"errors"
	"fmt"
)

type WorkflowError struct {
	IsRetryable bool   `json:"retryable"`       // Whether the error is retryable
	Err         string `json:"error,omitempty"` // Include error string if Err is present
}

// NewErrorWrapperWorkflowError creates a new NotFound error instance
func NewErrorWrapperWorkflowError(err error) *WorkflowError {
	if err == nil {
		return nil
	}
	if IsWorkflowError(err) {
		return err.(*WorkflowError)
	}
	return &WorkflowError{
		IsRetryable: true,
		Err:         err.Error(),
	}
}

// AsNonRetryableApplicationError returns a temporal ApplicationError flagged as non-retryable
func (s *WorkflowError) AsNonRetryableApplicationError() error {
	s.IsRetryable = false
	return s
}

// NewWorkflowError is a helper function to create a retryable CustomError.
func NewWorkflowError(underlyingErr error, retryAble bool) error {
	return &WorkflowError{
		IsRetryable: retryAble,
		Err:         underlyingErr.Error(),
	}
}

// Error method to satisfy the error interface.
func (e *WorkflowError) Error() string {
	retryableStatus := "non-retryable"
	if e.IsRetryable {
		retryableStatus = "retryable"
	}
	if len(e.Err) > 0 {
		return fmt.Sprintf("Error: (retryable: %s), underlying error: %s", retryableStatus, e.Err)
	}
	return fmt.Sprintf("Internal (retryable: %s)", retryableStatus)
}

// Unwrap method to support error wrapping (Go 1.13+).
// This allows you to use errors.Is and errors.As with the underlying error.
func (e *WorkflowError) Unwrap() error {
	return errors.New(e.Err)
}

func ToWorkflowError(err error) (*WorkflowError, bool) {
	var workflowError *WorkflowError
	ok := errors.As(err, &workflowError)
	return workflowError, ok
}

// IsRetryableError checks if the given error is a WorkflowError and if it's retryable.
func IsRetryableError(err error) bool {
	workflowError, ok := ToWorkflowError(err)
	if ok {
		return workflowError.IsRetryable
	}
	return false // Not a WorkflowError
}

func IsWorkflowError(err error) bool {
	_, ok := ToWorkflowError(err)
	return ok
}
