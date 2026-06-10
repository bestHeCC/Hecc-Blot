package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	iCoreApi "hecc-blot/contract/api"
	iCoreCache "hecc-blot/contract/cache"
	iCoreDb "hecc-blot/contract/db"
	iCoreError "hecc-blot/contract/error"
	iCoreLog "hecc-blot/contract/log"
	iCoreSse "hecc-blot/contract/sse"
	entityApi "hecc-blot/entity/api"
	coreConfig "hecc-blot/entity/config"
	"hecc-blot/enum/response"
	"hecc-blot/service/api"
	"hecc-blot/service/cache"
	"hecc-blot/service/db"
	errorSvc "hecc-blot/service/error"
	"hecc-blot/service/ioc"
	"hecc-blot/service/log"
	"hecc-blot/service/sse"
	"hecc-blot/service/trace"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/plugin/soft_delete"
)

func main() {
	// 用于收集所有出现的错误
	var allErrors []error

	// 加载配置
	config, err := initConf("/config.yaml")
	if err != nil {
		allErrors = append(allErrors, err)
	}

	traceSvc, traceClearUp, err := trace.NewTraceSvc(&config.Trace)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	logSvc, err := log.NewLogger(&config.Log)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	dbFactory, dbClearUp, err := db.NewDbFactory(&config.Db, logSvc)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
	responseSvc := api.NewResponseSvc()

	// 如果有任何错误发生，统一进行处理
	if len(allErrors) > 0 {
		panic(fmt.Errorf("以下错误发生：%v", allErrors))
	}

	defer func() {
		if dbFactory != nil {
			dbClearUp()
		}
		if traceSvc != nil {
			traceClearUp()
		}
		if cacheFactory.Redis() != nil {
			cacheFactory.Redis().Close()
		}
	}()

	// 注册至ioc容器
	ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
	ioc.Set(new(iCoreLog.ILog), logSvc)
	ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
	ioc.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc)
	register(apiHandle)

	sseHandle := sse.NewSseSvc(apiHandle.Engine())
	registerSse(sseHandle)

	apiHandle.Listen()
}

// initConf 加载统一配置
func initConf(configPath string) (*coreConfig.Config, error) {
	var config *coreConfig.Config

	wd, _ := os.Getwd()
	configPath = fmt.Sprintf("%s%s", wd, configPath)

	v := viper.New()
	v.SetConfigFile(configPath)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// 加载配置
	if err = v.Unmarshal(&config); err != nil {
		return nil, err
	}
	fmt.Println(fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", 34, fmt.Sprintf("加载配置成功，配置是%+v", config)))

	return config, nil
}

// register 注册api
func register(apiHandle iCoreApi.IApiHandle) {
	// iCoreApi.IApiHandle的Middleware方法，完成自动注入
	apiHandle.Middleware(&ReplayMiddleware{}, &TokenMiddleware{})
	{
		// iCoreApi.IApiHandle的Post方法（Get方法同理），完成自动注入
		// 自动校验参数，同时自动包装返回值
		// 统一返回值为{code:200, message:"请求成功", data:{}}
		apiHandle.Post("example/api", &ExampleApi{})
	}
}

func registerSse(sseHandle iCoreSse.ISseHandle) {
	sseHandle.Get("example/sse", &ExampleSse{})
}

// ReplayMiddleware 定义Replay中间件
type ReplayMiddleware struct {
	CacheFactory iCoreCache.ICacheFactory
}

func (r ReplayMiddleware) Middleware() interface{} {
	return func(c *gin.Context) {
		c.Next()
	}
}

// TokenMiddleware 定义Token中间件
type TokenMiddleware struct {
	ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() interface{} {
	return func(c *gin.Context) {
		// 从请求头中获取id
		u, err := strconv.ParseUint(c.GetHeader("id"), 10, 64) // 以10进制解析为int64
		if err != nil {
			t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, err))
			c.Abort()
		}

		c.Set("id", u)
		c.Next()
	}
}

// AccountModel 定义model
type AccountModel struct {
	ID          int                   `json:"id" gorm:"primaryKey"`
	AccountName string                `json:"account_name"`
	Password    string                `json:"password"`
	CreatedAt   int64                 `json:"created_at"`
	UpdatedAt   int64                 `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"deleted_at"`
}

func (b AccountModel) TableName() string {
	return "account"
}

// GetID model需要实现iCoreDb.IModel接口
func (b AccountModel) GetID() int {
	return b.ID
}

// ExampleRequest 定义Add请求参数
type ExampleRequest struct {
	Name string `json:"name" binding:"required"`
}

// GetMessages 自定义错误信息
func (a ExampleRequest) GetMessages() entityApi.Messages {
	return entityApi.Messages{
		"Name.required": "名字不能为空",
	}
}

// ExampleApi 定义接口
type ExampleApi struct {
	// 注意结构体内的字段需要保证顺序，注入的服务需要放在最前面，请求参数需要放在最后面
	// 通过inject tag，注册路由时会自动注入对应服务
	DbFactory    iCoreDb.IDbFactory       `inject:""`
	LogSvc       iCoreLog.ILog            `inject:""`
	CacheFactory iCoreCache.ICacheFactory `inject:""`

	// 请求参数，注册路由时会自动绑定并校验
	ExampleRequest
}

func (e ExampleApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	data := AccountModel{
		AccountName: "example",
	}

	e.LogSvc.Info(ctx, "log", "account", data)

	// e.DbFactory.Build(ctx) 返回一个iCoreDb.IDb，用于操作数据库，默认使用mysql
	mysqlSvc := e.DbFactory.Build(ctx)
	// 开启事务
	tx := mysqlSvc.Begin()
	err := tx.Where("id = ?", 2).Add(&data)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	// 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 使用缓存
	e.CacheFactory.Local().Set(ctx, "data", data, 10)
	e.CacheFactory.Redis().Set(ctx, "data", data, -1)

	// 此处只需要关注返回值，接口返回格式由iCoreApi.IResponse统一处理
	return data, nil
}

// ExampleSse 定义sse接口
type ExampleSse struct {
	LogSvc iCoreLog.ILog `inject:""`
}

func (e ExampleSse) Serve(ctx *gin.Context) error {
	e.LogSvc.Info(ctx, "sse start")

	// 获取流写入对象
	writer := ctx.Writer
	// 确保连接关闭时退出循环
	clientCtx := ctx.Request.Context()

	// 循环推送消息（模拟实时数据）
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientCtx.Done():
			// 客户端断开连接，退出
			fmt.Println("客户端断开连接")
			return nil
		case <-ticker.C:
			// 构造 SSE 消息（固定格式）
			msg := fmt.Sprintf("data: 当前服务器时间：%s\n\n", time.Now().Format(time.RFC3339))
			// 写入响应流
			_, err := writer.WriteString(msg)
			if err != nil {
				return err
			}
			// 立即刷新到客户端
			writer.Flush()
		}
	}
}
