package response

// Error is used to pass an error during the request through the
// application with web specific context.
type Error struct {
	Err    error
	Status int
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *Error) Error() string {
	return re.Err.Error()
}

// NewError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewError(err error, status int) error {
	return &Error{err, status}
}