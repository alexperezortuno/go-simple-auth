package errors

type ErrorCode int

const (
	TokenInvalidOrExpired ErrorCode = iota - 2001
	FailedToGenerateToken
	InvalidToken
	TokenRequired
	InvalidFormat ErrorCode = iota - 3001
	InvalidUserOrPassword
	RateLimitExceeded   ErrorCode = iota - 4001
	InternalServerError ErrorCode = iota - 5001
)

var errorMessages = map[ErrorCode]string{
	TokenInvalidOrExpired: "token is invalid or expired",
	FailedToGenerateToken: "failed to generate token",
	InvalidToken:          "invalid token",
	TokenRequired:         "token required",
	InvalidFormat:         "invalid format",
	InvalidUserOrPassword: "invalid user or password",
	RateLimitExceeded:     "rate limit exceeded",
	InternalServerError:   "internal server error",
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

func GetErrorMessage(code ErrorCode) string {
	return errorMessages[code]
}
