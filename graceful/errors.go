package graceful

import "net/http"

type (
	//UnauthorizedError - class needed to handle unauthorized cases gracefully, separate from other serious errors.
	UnauthorizedError struct {
		Message string
	}
	//BadRequestError - specific error class for bar requests handing
	BadRequestError struct {
		Message string
	}
	//StatusCode - interface for getting http status code from errors
	StatusCode interface {
		StatusCode() int
	}
)

//Error - function required by error interface; It returns default message if not defined.
func (gu UnauthorizedError) Error() string {
	if gu.Message != "" {
		return gu.Message
	}
	return "unauthorized"
}

//StatusCode - function to meet StatusCode interface - returns Unauthorized http code
func (gu UnauthorizedError) StatusCode() int {
	return http.StatusUnauthorized
}

//Error - function required by error interface; It returns default message if not defined.
func (br BadRequestError) Error() string {
	if br.Message != "" {
		return br.Message
	}
	return "badrequest"
}

//StatusCode - function to meet StatusCode interface - returns BadRequest http code
func (br BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}
