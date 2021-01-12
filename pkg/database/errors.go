package database

type Error string
func (e Error) Error() string { return string(e) }

const NotFound = Error("item could not be found")
