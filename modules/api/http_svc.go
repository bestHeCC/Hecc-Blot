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

	coreError "github.com/bestHeCC/hecc-api/error"
	iCoreApi "github.com/bestHeCC/hecc-core/contract/api"
	"github.com/bestHeCC/hecc-core/contract/ioc"
	iCoreTrace "github.com/bestHeCC/hecc-core/contract/trace"
	"github.com/bestHeCC/hecc-core/entity/config/server"
	envEnum "github.com/bestHeCC/hecc-core/enum/env"
	"github.com/bestHeCC/hecc-core/enum/response"
	"github.com/bestHeCC/hecc-core/util"

	"github.com/gin-gonic/gin"
)

type ApiHandle struct {
	config      *server.Config
	engine      *gin.Engine
	responseSvc iCoreApi.IResponse
	container   ioc.IContainer
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
		f.container.Inject(iMiddleware)

		middlewareValue := iMiddleware.Middleware()
		if middlewareValue != nil && reflect.TypeOf(middlewareValue).Kind() == reflect.Func {
			f.engine.Use(middlewareValue.(func(*gin.Context)))
		}
	}

	return f
}

func (f *ApiHandle) Get(apiPath string, apiInstance any) {
	f.registerAPI(apiPath, apiInstance, http.MethodGet)
}

func (f *ApiHandle) Post(apiPath string, apiInstance any) {
	f.registerAPI(apiPath, apiInstance, http.MethodPost)
}

func (f *ApiHandle) Listen() {
	readTimeout := f.config.ReadTimeout
	writeTimeout := f.config.WriteTimeout
	idleTimeout := f.config.IdleTimeout
	if readTimeout <= 0 {
		readTimeout = 30
	}
	if writeTimeout <= 0 {
		writeTimeout = 30
	}
	if idleTimeout <= 0 {
		idleTimeout = 60
	}
	srv := &http.Server{
		Addr:         ":" + f.config.Port,
		Handler:      f.engine,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
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

func (f *ApiHandle) registerAPI(apiPath string, apiInstance any, method string) {
	// 注入api依赖项（仅注入共享的服务依赖，请求参数字段留空）
	f.container.Inject(apiInstance)

	if _, ok := apiInstance.(iCoreApi.IApi); ok {
		// 缓存具体类型，避免每个请求都做类型断言
		apiType := reflect.TypeOf(apiInstance).Elem()

		handler := func(c *gin.Context) {
			// 每个请求创建独立实例，避免并发请求共享写入
			newInstance := reflect.New(apiType).Interface()
			f.container.Inject(newInstance)
			api := newInstance.(iCoreApi.IApi)

			// 自动绑定参数，并进行校验
			if err := c.ShouldBind(newInstance); err != nil {
				f.responseSvc.Regular(c, nil, coreError.New(response.ValidateError, util.GetErrorMsg(api, err)))
				return
			}

			resp, err := api.Call(c)
			f.responseSvc.Regular(c, resp, err)
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

func NewApiSvc(config *server.Config, responseSvc iCoreApi.IResponse, traceSvc iCoreTrace.ITrace, container ioc.IContainer) iCoreApi.IApiHandle {
	mode, ok := mapEnv[config.Env]
	if !ok {
		panic(fmt.Sprintf("无效环境配置:%s", mode))
	}

	if container == nil {
		panic("ioc: 注入容器不能为空")
	}

	gin.SetMode(mode)
	app := gin.New()
	app.Use(gin.Recovery())
	app.Use(bodySizeLimit(config.BodySizeLimit)) // 10MB

	apiHandle := &ApiHandle{
		config:      config,
		engine:      app,
		responseSvc: responseSvc,
		container:   container,
	}

	if traceSvc != nil {
		// 开启链路追踪
		traceMiddleware := &HttpTraceMiddleware{
			TraceSvc: traceSvc,
		}
		apiHandle.Middleware(traceMiddleware)
	}

	return apiHandle
}

// bodySizeLimit 限制请求体大小，防止大 payload 攻击
func bodySizeLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes == 0 {
		maxBytes = 10 << 20
	}
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
