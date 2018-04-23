package graceful

import "fmt"

type (
	DomainCode int
	Error struct {
		message  string
		code     int
		withCode bool
		domain   DomainCode
	}
)

const (
	VkDomain      DomainCode = iota + 1
	NetworkDomain
	RedisDomain
	MySqlDomain
	ParsingDomain
	FbDomain
	InvalidDomain
)

func (c DomainCode) String() string {
	switch c {
	case VkDomain:
		return "vk"
	case NetworkDomain:
		return "network"
	case RedisDomain:
		return "redis"
	case MySqlDomain:
		return "mysql"
	case ParsingDomain:
		return "parsing"
	case FbDomain:
		return "fb"
	case InvalidDomain:
		return "invalid"
	default:
		return "unknown"
	}
}

func (e Error) Error() string {
	if e.withCode {
		return fmt.Sprintf("d=%s; code=%d; message=%s", e.domain, e.code, e.message)
	} else {
		return fmt.Sprintf("d=%s; message=%s", e.domain, e.message)
	}
}

func (e Error) Code() (bool, int) {
	return e.withCode, e.code
}

func (e Error) Domain() DomainCode {
	return e.domain
}

func newError(domain DomainCode, message string, code ...int) *Error {
	switch len(code) {
	case 0:
		return &Error{message, 0, false, domain}
	default:
		return &Error{message, code[0], true, domain}
	}
}

func NewVkError(message string, code ...int) *Error {
	return newError(VkDomain, message, code...)
}

func NewNetworkError(message string, code ...int) *Error {
	return newError(NetworkDomain, message, code...)
}

func NewMySqlError(message string, code ...int) *Error {
	return newError(MySqlDomain, message, code...)
}

func NewRedisError(message string, code ...int) *Error {
	return newError(RedisDomain, message, code...)
}

func NewParsingError(message string, code ...int) *Error {
	return newError(ParsingDomain, message, code...)
}

func NewFbError(message string, code ...int) *Error {
	return newError(FbDomain, message, code...)
}

func NewInvalidError(message string, code ...int) *Error {
	return newError(InvalidDomain, message, code...)
}
