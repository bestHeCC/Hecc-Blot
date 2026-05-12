package log

import (
	"context"
	"fmt"

	ilog "core/contract/log"
	"core/entity/config/log"
	"core/enum/trace"

	"github.com/gin-gonic/gin"
)

func getTraceId(ctx context.Context) string {
	traceId := ctx.Value(trace.TraceIdKey)
	if traceId != nil {
		return traceId.(string)
	}

	return ""
}

// extractContext 从 *gin.Context 中提取 context.Context
func extractContext(ctx context.Context) context.Context {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		return ginCtx.Request.Context()
	}
	return ctx
}

func NewLogger(config *log.Config) (ilog.ILog, error) {
	// 优先使用sls日志服务
	slsConfig := config.Sls
	if slsConfig.Enable {
		return newSlsSvc(&slsConfig)
	}

	// sls日志服务未开启，则使用本地日志服务
	localConfig := config.Local
	if localConfig.Enable {
		return newLogSvc(&localConfig)
	}

	return nil, fmt.Errorf("未配置日志服务")
}
