package main

type Error struct {
	str string
}

func (e *Error) Error() string {
	return e.str
}

func NewError(str string) *Error {
	return &Error{str}
}
