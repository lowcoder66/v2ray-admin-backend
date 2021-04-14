package response

type ErrorResponseBody struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}
type MsgResponseBody struct {
	Message string `json:"message"`
}

func ErrRes(message string, error interface{}) ErrorResponseBody {
	return ErrorResponseBody{message, error}
}
func MessageRes(message string) MsgResponseBody {
	return MsgResponseBody{message}
}
