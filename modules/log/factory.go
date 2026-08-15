package log

import (
	"context"
	"fmt"

	iLog "github.com/bestHeCC/hecc-core/contract/log"
	"github.com/bestHeCC/hecc-core/entity/config/log"
	"github.com/bestHeCC/hecc-core/enum/trace"
)

func getTraceId(ctx context.Context) string {
	traceId := ctx.Value(trace.TraceIdKey)
	if traceId != nil {
		return traceId.(string)
	}

	return ""
}

func NewLogger(config *log.Config) (iLog.ILog, error) {
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
