package api

import (
	"net/http"

	iCoreApi "github.com/bestHeCC/hecc-core/contract/api"
	"github.com/bestHeCC/hecc-core/contract/ratelimit"
	"github.com/bestHeCC/hecc-core/enum/response"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware 请求频率限流中间件，按客户端 IP 限流。
type RateLimitMiddleware struct {
	Limiter ratelimit.RateLimiter
}

// NewRateLimitMiddleware 创建限流中间件。
// limiter 由调用方注入，可在内存与 Redis 后端之间选择（见 NewMemoryLimiter / hecc-cache 的 NewRedisLimiter）。
func NewRateLimitMiddleware(limiter ratelimit.RateLimiter) iCoreApi.IMiddleware {
	if limiter == nil {
		panic("api: 限流器不能为空")
	}
	return &RateLimitMiddleware{Limiter: limiter}
}

func (r *RateLimitMiddleware) Middleware() any {
	return func(c *gin.Context) {
		if !r.Limiter.Allow(c.Request.Context(), c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    response.RateLimit,
				"message": response.CodeMap[response.RateLimit],
				"data":    nil,
			})
			return
		}
		c.Next()
	}
}
