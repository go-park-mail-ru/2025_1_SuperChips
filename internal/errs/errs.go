package errs


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

var (
	ErrForbidden  = &StatusCodeError{Code: 403, Msg: "invalid credentials"}
	ErrValidation = &StatusCodeError{Code: 400, Msg: "validation failed"}
	ErrConflict   = &StatusCodeError{Code: 409, Msg: "resource conflict"}
	ErrNotFound   = &StatusCodeError{Code: 404, Msg: "resource not found"}
	ErrInternal   = &StatusCodeError{Code: 500, Msg: "internal server error"}
)