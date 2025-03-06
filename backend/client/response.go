package client

import "fmt"

type Response struct {
	Code    int
	Message string
	Data    interface{}
}

func (a *App) errorResponse(message ...interface{}) Response {
	return Response{
		Code:    400,
		Message: fmt.Sprint(message...),
		Data:    nil,
	}
}

func (a *App) successResponse(msg string, data interface{}) Response {
	return Response{
		Code:    200,
		Message: msg,
		Data:    data,
	}
}
