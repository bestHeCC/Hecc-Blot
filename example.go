package main

import (
	"fmt"
	"os"
	"strconv"

	iCoreApi "core/contract/api"
	iCoreCache "core/contract/cache"
	iCoreDb "core/contract/db"
	iCoreError "core/contract/error"
	iCoreLog "core/contract/log"
	entityApi "core/entity/api"
	coreConfig "core/entity/config"
	"core/enum/response"
	"core/service/api"
	"core/service/cache"
	"core/service/db"
	errorSvc "core/service/error"
	"core/service/ioc"
	"core/service/log"

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

	logSvc, err := log.NewLogger(config)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	dbFactory, deClearUp, err := db.NewDbFactory(config, logSvc)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	cacheFactory := cache.NewCacheFactory(&config.Cache)
	responseSvc := api.NewResponseSvc()

	// 如果有任何错误发生，统一进行处理
	if len(allErrors) > 0 {
		panic(fmt.Errorf("以下错误发生：%v", allErrors))
	}

	defer func() {
		if dbFactory != nil {
			deClearUp()
		}
	}()

	// 注册至ioc容器
	ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
	ioc.Set(new(iCoreLog.ILog), logSvc)
	ioc.Set(new(iCoreCache.ICacheFactory), cacheFactory)
	ioc.Set(new(iCoreApi.IResponse), responseSvc)

	apiHandle := api.NewApiSvc(config, responseSvc)
	register(apiHandle)
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
		apiHandle.Post("account/add", &AddApi{})
	}
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

// AddRequest 定义Add请求参数
type AddRequest struct {
	Name string `json:"name" binding:"required"`
}

// GetMessages 自定义错误信息
func (a AddRequest) GetMessages() entityApi.Messages {
	return entityApi.Messages{
		"Name.required": "名字不能为空",
	}
}

// AddApi 定义Add接口
type AddApi struct {
	// 注意结构体内的字段需要保证顺序，注入的服务需要放在最前面，请求参数需要放在最后面
	// 通过inject tag，注册路由时会自动注入对应服务
	DbFactory iCoreDb.IDbFactory `inject:""`
	LogSvc    iCoreLog.ILog      `inject:""`

	// 请求参数，注册路由时会自动绑定并校验
	AddRequest
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	newAccount := AccountModel{
		AccountName: a.Name, // 请求参数自动绑定
	}

	a.LogSvc.Info(ctx, "add account", "account", newAccount)

	// a.DbFactory.Build(ctx) 返回一个iCoreDb.IDb，用于操作数据库，默认使用mysql
	err := a.DbFactory.Build(ctx).Add(&newAccount)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 此处只需要关注返回值，接口返回格式由iCoreApi.IResponse统一处理
	return newAccount, nil
}
