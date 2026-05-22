package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	iCoreApi "hecc-blot/contract/api"
	iCoreTrace "hecc-blot/contract/trace"
	"hecc-blot/entity/config/server"
	envEnum "hecc-blot/enum/env"
	"hecc-blot/enum/response"
	coreError "hecc-blot/service/error"
	"hecc-blot/service/ioc"
	"hecc-blot/service/trace"

	"github.com/gin-gonic/gin"
)

type ApiHandle struct {
	config      *server.Config
	engine      *gin.Engine
	responseSvc iCoreApi.IResponse
}

var mapEnv = map[envEnum.Value]string{
	envEnum.DevMode:     gin.DebugMode,
	envEnum.ProductMode: gin.ReleaseMode,
	envEnum.TestMode:    gin.TestMode,
}

func (f *ApiHandle) Engine() *gin.Engine {
	return f.engine
}

func (f *ApiHandle) Middleware(middlewares ...iCoreApi.IMiddleware) iCoreApi.IApiHandle {
	for _, iMiddleware := range middlewares {
		ioc.Inject(iMiddleware)

		middlewareValue := iMiddleware.Middleware()
		if middlewareValue != nil && reflect.TypeOf(middlewareValue).Kind() == reflect.Func {
			f.engine.Use(middlewareValue.(func(*gin.Context)))
		}
	}

	return f
}

func (f *ApiHandle) Get(apiPath string, apiInstance interface{}) {
	f.registerAPI(apiPath, apiInstance, http.MethodGet)
}

func (f *ApiHandle) Post(apiPath string, apiInstance interface{}) {
	f.registerAPI(apiPath, apiInstance, http.MethodPost)
}

func (f *ApiHandle) Listen() {
	srv := &http.Server{
		Addr:    ":" + f.config.Port,
		Handler: f.engine,
	}

	// 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为重启做准备
	quit := make(chan os.Signal, 1)
	// 接收SIGINT（Ctrl + C）和SIGTERM信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// 创建一个5秒的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 等待所有请求完成处理后关闭服务
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exiting")
}

func (f *ApiHandle) registerAPI(apiPath string, apiInstance interface{}, method string) {
	// 注入api依赖项
	ioc.Inject(apiInstance)

	if v, ok := apiInstance.(iCoreApi.IApi); ok {
		handler := func(c *gin.Context) {
			// 自动绑定参数，并进行校验
			if err := c.ShouldBind(&v); err != nil {
				f.responseSvc.Regular(c, nil, coreError.Newf(response.ValidateError, GetErrorMsg(v, err)))

				return
			}

			resp, err := v.Call(c)
			f.responseSvc.Regular(c, resp, err)

			return
		}

		switch method {
		case http.MethodGet:
			f.engine.GET(apiPath, handler)
		case http.MethodPost:
			f.engine.POST(apiPath, handler)
		default:
			panic(
				fmt.Sprintf("无效http请求类型，%s", method),
			)
		}
	}
}

func NewApiSvc(config *server.Config, responseSvc iCoreApi.IResponse, traceSvc iCoreTrace.ITrace) iCoreApi.IApiHandle {
	mode, ok := mapEnv[config.Env]
	if !ok {
		panic(fmt.Sprintf("无效环境配置:%s", mode))
	}

	gin.SetMode(mode)
	app := gin.New()

	apiHandle := &ApiHandle{
		config:      config,
		engine:      app,
		responseSvc: responseSvc,
	}

	if traceSvc != nil {
		// 开启链路追踪
		traceMiddleware := &trace.HttpTraceMiddleware{
			TraceSvc: traceSvc,
		}
		apiHandle.Middleware(traceMiddleware)
	}

	return apiHandle
}
