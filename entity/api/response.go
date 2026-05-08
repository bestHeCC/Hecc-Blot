package api

import (
	"core/enum/response"
)

type responseBody struct {
	Code    response.Value `json:"code"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
}
