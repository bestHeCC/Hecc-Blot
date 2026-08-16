package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bestHeCC/hecc-core/contract/ratelimit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewMemoryLimiter(ratelimit.Config{Algorithm: ratelimit.SlidingWindow, Limit: 1, Window: 60})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(NewRateLimitMiddleware(limiter).Middleware().(func(*gin.Context)))
	engine.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 首次放行
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	// 同 IP 第二次超限，返回 429 + 统一响应格式
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), "请求过于频繁")
}
