package main

type Error struct {
	str string
}

func (e *Error) Error() string {
	return e.str
}

func error(str string) *Error {
	return &Error{str}
}
