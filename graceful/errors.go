package graceful

import "fmt"

type (
	DomainCode int
	Error struct {
		message  string
		codes    []int
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
	if len(e.codes) > 0 {
		return fmt.Sprintf("d=%s; c=%v; m=%s", e.domain, e.codes, e.message)
	} else {
		return fmt.Sprintf("d=%s; m=%s", e.domain, e.message)
	}
}

func (e Error) Code() (bool, int) {
	if len(e.codes) > 0 {
		return true, e.codes[0]
	} else {
		return false, 0
	}
}

func (e Error) Codes() []int {
	return e.codes
}

func (e Error) Domain() DomainCode {
	return e.domain
}

func newError(domain DomainCode, message string, codes ...int) *Error {
	return &Error{message, codes, domain}
}

func NewVkError(message string, codes ...int) *Error {
	return newError(VkDomain, message, codes...)
}

func NewNetworkError(message string, codes ...int) *Error {
	return newError(NetworkDomain, message, codes...)
}

func NewMySqlError(message string, codes ...int) *Error {
	return newError(MySqlDomain, message, codes...)
}

func NewRedisError(message string, codes ...int) *Error {
	return newError(RedisDomain, message, codes...)
}

func NewParsingError(message string, codes ...int) *Error {
	return newError(ParsingDomain, message, codes...)
}

func NewFbError(message string, codes ...int) *Error {
	return newError(FbDomain, message, codes...)
}

func NewInvalidError(message string, codes ...int) *Error {
	return newError(InvalidDomain, message, codes...)
}
