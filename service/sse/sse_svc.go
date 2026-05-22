package sse

import (
	"net/http"
	"reflect"

	iCoreApi "hecc-blot/contract/api"
	iCoreSse "hecc-blot/contract/sse"
	"hecc-blot/service/ioc"

	"github.com/gin-gonic/gin"
)

type SseHandle struct {
	engine *gin.Engine
}

func (f *SseHandle) Middleware(middlewares ...iCoreApi.IMiddleware) iCoreSse.ISseHandle {
	for _, iMiddleware := range middlewares {
		ioc.Inject(iMiddleware)

		middlewareValue := iMiddleware.Middleware()
		if middlewareValue != nil && reflect.TypeOf(middlewareValue).Kind() == reflect.Func {
			f.engine.Use(middlewareValue.(func(*gin.Context)))
		}
	}

	return f
}

func (f *SseHandle) Get(apiPath string, sseInstance iCoreSse.ISse) {
	ioc.Inject(sseInstance)

	f.engine.GET(apiPath, func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		if err := sseInstance.Serve(c); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
	})
}

func NewSseSvc(engine *gin.Engine) iCoreSse.ISseHandle {
	return &SseHandle{
		engine: engine,
	}
}
