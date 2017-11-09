package common

func NewErrorWithCode(code int64, msg string) error {
	return &ErrorWithCode{Code: code, Msg: msg}
}

type ErrorWithCode struct {
	Code int64
	Msg  string
}

func (e *ErrorWithCode) Error() string {
	return e.Msg
}
