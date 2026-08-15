package sse

import iCoreApi "github.com/bestHeCC/hecc-core/contract/api"

type ISseHandle interface {
	Get(apiPath string, sse ISse)
	Post(apiPath string, sse ISse)
	Middleware(middlewares ...iCoreApi.IMiddleware) ISseHandle
	Group(relativePath string, middlewares ...iCoreApi.IMiddleware) ISseHandle
}
