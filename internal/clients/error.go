package clients

// ErrorRetryable represents retryable type error.
type ErrorRetryable struct {
	Err error
}

func (e ErrorRetryable) Error() string {
	return e.Err.Error()
}

// ErrorWrongPassword represents wrong password type error.
type ErrorWrongPassword struct {
	Err error
}

func (e ErrorWrongPassword) Error() string {
	return e.Err.Error()
}
