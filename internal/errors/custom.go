package errors

type ErrorCode int

const (
	TokenInvalidOrExpired ErrorCode = iota - 2001
	FailedToGenerateToken
	InvalidFormat ErrorCode = iota - 3001
	InvalidUserOrPassword
)

var errorMessages = map[ErrorCode]string{
	TokenInvalidOrExpired: "token is invalid or expired",
	FailedToGenerateToken: "failed to generate token",
	InvalidFormat:         "invalid format",
	InvalidUserOrPassword: "invalid user or password",
}

type CustomError struct {
	Message string    `json:"message"`
	Code    ErrorCode `json:"code"`
}

func NewCustomError(code ErrorCode) *CustomError {
	return &CustomError{
		Message: errorMessages[code],
		Code:    code,
	}
}
