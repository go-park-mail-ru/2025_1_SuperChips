package error


type StatusError interface {
	error
	StatusCode() int
}

type StatusCodeError struct {
	Code int
	Msg  string
}

func (e *StatusCodeError) Error() string {
	return e.Msg
}

func (e *StatusCodeError) StatusCode() int {
	return e.Code
}
