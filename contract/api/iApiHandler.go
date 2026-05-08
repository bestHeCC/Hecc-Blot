package api

type IApiHandle interface {
	Get(apiPath string, api interface{})
	Listen()
	Middleware(middlewares ...IMiddleware) IApiHandle
	Post(apiPath string, api interface{})
}
