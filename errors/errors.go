package errors

import "fmt"

// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type errBadRequest struct {
	message string
}

func (err *errBadRequest) Error() string {
	if err == nil {
		return "nil"
	}
	return err.message
}

// NewBadRequestError create a new not found error
func NewBadRequestError(message string) error {
	return &errBadRequest{message}
}

// IsBadRequestError judges error is errBadRequest
func IsBadRequestError(err error) bool {
	if _, ok := err.(*errBadRequest); ok {
		return true
	}
	return false
}

type errNotFound struct {
	resource string
}

func (err *errNotFound) Error() string {
	if err == nil {
		return "nil"
	}
	return fmt.Sprintf("resource '%s' is not found", err.resource)
}

// NewNotFoundError create a new not found error
func NewNotFoundError(resource string) error {
	return &errNotFound{resource}
}

// IsNotFoundError judges error is errNotFound
func IsNotFoundError(err error) bool {
	if _, ok := err.(*errNotFound); ok {
		return true
	}
	return false
}

type errConflict struct {
	resource string
}

func (err *errConflict) Error() string {
	if err == nil {
		return "nil"
	}
	return fmt.Sprintf("resource '%s' is conflict with the exists one", err.resource)
}

// NewConflictError create a new conflict error
func NewConflictError(resource string) error {
	return &errConflict{resource}
}

// IsConflictError judges error is errConflict
func IsConflictError(err error) bool {
	if _, ok := err.(*errConflict); ok {
		return true
	}
	return false
}

// not ready will returen StatusNotAcceptable http status if user  post or put a request
type errNotReady struct {
	resource string
}

func (err *errNotReady) Error() string {
	if err == nil {
		return "nil"
	}
	return fmt.Sprintf("resource '%s' is not ready for process", err.resource)
}

// NewNotReadyError create a new not ready error
func NewNotReadyError(resource string) error {
	return &errNotReady{resource}
}

// IsNotReadyError judges error is errNotReady
func IsNotReadyError(err error) bool {
	if _, ok := err.(*errNotReady); ok {
		return true
	}
	return false
}

type errTaskIsRunning struct {
	message string
}

func (err *errTaskIsRunning) Error() string {
	if err == nil {
		return "nil"
	}
	return err.message
}

// NewTaskIsRunningError create a new task is running error
func NewTaskIsRunningError(message string) error {
	return &errTaskIsRunning{message}
}

// IsTaskIsRunningError judges error is errTaskIsRunning
func IsTaskIsRunningError(err error) bool {
	if _, ok := err.(*errTaskIsRunning); ok {
		return true
	}
	return false
}

// NewClientError return a ClientError with the msg
func NewClientError(msg string) *ClientError {
	return &ClientError{message: msg}
}

// ClientError is an error caused by request client
type ClientError struct {
	message string
}

func (err *ClientError) Error() string {
	return err.message
}

// IsClientError judeges error is IsClientError
func IsClientError(err error) bool {
	if _, ok := err.(*ClientError); ok {
		return true
	}
	return false
}

// NewServerError return a ServerError with the msg
func NewServerError(msg string) *ServerError {
	return &ServerError{message: msg}
}

// ServerError is an error caused by request client
type ServerError struct {
	message string
}

func (err *ServerError) Error() string {
	return err.message
}

// invalid region
type errInvalidRegion struct {
	region string
}

func (err *errInvalidRegion) Error() string {
	if err == nil {
		return "nil"
	}
	return fmt.Sprintf("invalid region '%s'", err.region)
}

// NewInvalidRegionError create a new not found error
func NewInvalidRegionError(region string) error {
	return &errInvalidRegion{region}
}

// IsInvalidRegionError judges error is errInvalidRegion
func IsInvalidRegionError(err error) bool {
	if _, ok := err.(*errInvalidRegion); ok {
		return true
	}
	return false
}

type errForbidden struct {
	message string
}

func (err *errForbidden) Error() string {
	if err == nil {
		return "nil"
	}
	return err.message
}

// NewForbiddenError create a new not found error
func NewForbiddenError(message string) error {
	return &errForbidden{message}
}

// IsForbiddenError judges error is errForbidden
func IsForbiddenError(err error) bool {
	if _, ok := err.(*errForbidden); ok {
		return true
	}
	return false
}
