// Package fault describes operation failures without an HTTP dependency.
package fault

type Kind uint8

const (
	Invalid Kind = iota
	Unauthenticated
	Forbidden
	NotFound
	Conflict
	Unavailable
	RateLimited
	Unsupported
	TooLarge
)

type Error struct {
	Kind Kind
	Code string
}

func (e *Error) Error() string         { return e.Code }
func New(kind Kind, code string) error { return &Error{Kind: kind, Code: code} }
