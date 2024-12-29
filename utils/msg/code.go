package msg

const (
	CodeOK         = 200
	CodeBadRequest = 400
	CodeError      = 500
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
}

func Resp(code int, message string, data interface{}, requestID string) Response {
	return Response{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	}
}
