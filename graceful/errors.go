package graceful

type (
	//UnauthorizedError - class needed to handle unauthorized cases gracefully, separate from other serious errors.
	UnauthorizedError struct {
	}
)

//Error - function required by error interface; It returns default message.
func (gu UnauthorizedError) Error() string {
	return "unauthorized"
}
