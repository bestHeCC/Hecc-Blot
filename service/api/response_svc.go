package api

import (
	"context"
	"net/http"
	"sync"

	"hecc-blot/contract/api"
	coreError "hecc-blot/contract/error"
	"hecc-blot/enum/response"

	"github.com/gin-gonic/gin"
)

type ResponseSvc struct {
	Code    response.Value `json:"code"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
}

var responsePool = sync.Pool{
	New: func() interface{} { return &ResponseSvc{} },
}

func (r *ResponseSvc) Regular(ctx context.Context, data interface{}, err coreError.IError) {
	g := ctx.(*gin.Context)
	code := response.Success
	if err != nil {
		code = err.GetCode()
		data = err.GetData()
	}

	resp := responsePool.Get().(*ResponseSvc)
	defer responsePool.Put(resp)

	resp.Code = code
	resp.Message = response.CodeMap[code]
	resp.Data = data

	g.JSON(http.StatusOK, resp)
}

func NewResponseSvc() api.IResponse {
	return &ResponseSvc{}
}
