package social

import "fmt"

type (
	messageCodes struct {
		message string
		codes   []int
	}
	//VkError - type for storing errors returned by vk api (except codes fires unauthorized errors)
	VkError struct {
		messageCodes
	}
	//FbError - type for storing erros returned by fb api (except codes fires unauthorized errors)
	FbError struct {
		messageCodes
	}
	//UnauthorizedError - class needed to handle unauthorized cases gracefully, separate from other serious errors.
	UnauthorizedError struct {
	}
)

//NewVkError - construct VkError in comfort way
func NewVkError(message string, codes ...int) error {
	return &VkError{messageCodes{message, codes}}
}

//NewFbError - construct VkError in comfort way
func NewFbError(message string, codes ...int) error {
	return &FbError{messageCodes{message, codes}}
}

//Error - function required by error interface; It returns default message.
func (gu UnauthorizedError) Error() string {
	return "unauthorized"
}

//Error - function required be standard error interface. It is common fow some classes in the module.
func (mc messageCodes) Error() string {
	if len(mc.codes) > 0 {
		return fmt.Sprintf("c=%v; m=%s", mc.codes, mc.message)
	}
	return mc.message
}

//Codes - returns all codes of the current error object
func (mc messageCodes) Codes() []int {
	return mc.codes
}
