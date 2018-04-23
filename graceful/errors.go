package graceful

import "fmt"

type (
	DomainCode string
	Error struct {
		message  string
		code     int
		withCode bool
		domain   DomainCode
	}
)

const (
	DefaultDomain DomainCode = "default"
	VkDomain      DomainCode = "vk"
	NetworkDomain DomainCode = "network"
	RedisDomain   DomainCode = "redis"
	MySqlDomain   DomainCode = "mysql"
	ParsingDomain DomainCode = "parsing"
)

func (e Error) Error() string {
	if e.domain != DefaultDomain {
		if e.withCode {
			return fmt.Sprintf("%s > code:%d; message:%s", e.domain, e.code, e.message)
		} else {
			return fmt.Sprintf("%s > message:%s", e.domain, e.message)
		}
	} else {
		if e.withCode {
			return fmt.Sprintf("code:%d; message:%s", e.code, e.message)
		} else {
			return e.message
		}
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

func NewError(message string, code ...int) *Error {
	return newError(DefaultDomain, message, code...)
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
