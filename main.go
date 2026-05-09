package main

import (
	"fmt"
	"os"

	"core/api/account"
	iCoreApi "core/contract/api"
	iCoreCache "core/contract/cache"
	iCoreDb "core/contract/db"
	iCoreLog "core/contract/log"
	coreConfig "core/entity/config"
	"core/middleware"
	"core/service/api"
	"core/service/cache"
	"core/service/db"
	"core/service/ioc"
	"core/service/log"

	"github.com/spf13/viper"
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

func initConf(configPath string) (*coreConfig.Config, error) {
	var config *coreConfig.Config

	// 加载统一配置
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

func register(apiHandle iCoreApi.IApiHandle) {
	apiHandle.Middleware(&middleware.ReplayMiddleware{}, &middleware.TokenMiddleware{})
	{
		apiHandle.Post("account/add", &account.AddApi{})
	}
}
