package graceful

import "fmt"

type (
	//DomainCode - type for different error reasons (kinds) enumeration
	DomainCode int
	//Error - our class to store error reason and codes not just message
	Error struct {
		message string
		codes   []int
		domain  DomainCode
	}
)

const (
	//VkDomain - marks error from vk api response
	VkDomain DomainCode = iota + 1
	//NetworkDomain - marks error network error (cant resolve host for example)
	NetworkDomain
	//RedisDomain - error from redis driver
	RedisDomain
	//MySQLDomain - error from mysql driver
	MySQLDomain
	//ParsingDomain - error with parsing json or other formats
	ParsingDomain
	//FbDomain - marks error from fb api response
	FbDomain
	//InvalidDomain - marks error of unexpected format of response of third party services or other unexpected situations
	InvalidDomain
	//NotFoundDomain - marks errors of invalid third party tokens on registration and login
	NotFoundDomain
)

func (c DomainCode) String() string {
	switch c {
	case VkDomain:
		return "vk"
	case NetworkDomain:
		return "network"
	case RedisDomain:
		return "redis"
	case MySQLDomain:
		return "mysql"
	case ParsingDomain:
		return "parsing"
	case FbDomain:
		return "fb"
	case InvalidDomain:
		return "invalid"
	case NotFoundDomain:
		return "notfound"
	default:
		return "unknown"
	}
}

func (e Error) Error() string {
	if len(e.codes) > 0 {
		return fmt.Sprintf("d=%s; c=%v; m=%s", e.domain, e.codes, e.message)
	}
	return fmt.Sprintf("d=%s; m=%s", e.domain, e.message)
}

//Code - returns bool - is any code presented in error object & int - first code from array of codes of the error object
func (e Error) Code() (bool, int) {
	if len(e.codes) > 0 {
		return true, e.codes[0]
	}
	return false, 0
}

//Codes - method returns all codes error contains
func (e Error) Codes() []int {
	return e.codes
}

//Domain - returns error reason (domain)
func (e Error) Domain() DomainCode {
	return e.domain
}

func newError(domain DomainCode, message string, codes ...int) *Error {
	return &Error{message, codes, domain}
}

//NewVkError - construct new error with vk domain
func NewVkError(message string, codes ...int) *Error {
	return newError(VkDomain, message, codes...)
}

//NewNetworkError - construct new error with network domain
func NewNetworkError(message string, codes ...int) *Error {
	return newError(NetworkDomain, message, codes...)
}

//NewMySQLError - construct new error with mysql domain
func NewMySQLError(message string, codes ...int) *Error {
	return newError(MySQLDomain, message, codes...)
}

//NewRedisError - construct new error with redis domain
func NewRedisError(message string, codes ...int) *Error {
	return newError(RedisDomain, message, codes...)
}

//NewParsingError - construct new error with parsing domain
func NewParsingError(message string, codes ...int) *Error {
	return newError(ParsingDomain, message, codes...)
}

//NewFbError - construct new error with fb domain
func NewFbError(message string, codes ...int) *Error {
	return newError(FbDomain, message, codes...)
}

//NewInvalidError - construct new error with invalid domain
func NewInvalidError(message string, codes ...int) *Error {
	return newError(InvalidDomain, message, codes...)
}

//NewNotFoundError - construct new error with NotFound domain
func NewNotFoundError(message string, codes ...int) *Error {
	return newError(NotFoundDomain, message, codes...)
}
