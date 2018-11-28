package graceful

import "net/http"

type (
	//UnauthorizedError - class needed to handle unauthorized cases gracefully, separate from other serious errors.
	UnauthorizedError struct {
		Message string
	}
	//BadRequestError - specific error class for bar requests handling
	BadRequestError struct {
		Message string
	}
	//NotFoundError - specific error class for not found object handling
	NotFoundError struct {
		Message string
	}
	//ForbiddenError - specific error class for forbidden object handling
	ForbiddenError struct {
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

//Error - function required by error interface; It returns default message if not defined.
func (fb ForbiddenError) Error() string {
	if fb.Message != "" {
		return fb.Message
	}
	return "forbidden"
}

//StatusCode - function to meet StatusCode interface - returns Unauthorized http code
func (fb ForbiddenError) StatusCode() int {
	return http.StatusForbidden
}

//StatusCode - function to meet StatusCode interface - returns BadRequest http code
func (br BadRequestError) StatusCode() int {
	return http.StatusBadRequest
}

//Error - function required by error interface; It returns default message if not defined.
func (nf NotFoundError) Error() string {
	if nf.Message != "" {
		return nf.Message
	}
	return "notfound"
}

//StatusCode - function to meet StatusCode interface - returns NotFound http code
func (nf NotFoundError) StatusCode() int {
	return http.StatusNotFound
}
