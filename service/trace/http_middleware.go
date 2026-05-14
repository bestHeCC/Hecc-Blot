package trace

import (
	"context"
	"fmt"

	iCoreTrace "hecc-blot/contract/trace"
	"hecc-blot/enum/trace"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/propagation"
	otelTrace "go.opentelemetry.io/otel/trace"
)

type HttpTraceMiddleware struct {
	TraceSvc iCoreTrace.ITrace
}

func (h *HttpTraceMiddleware) Middleware() interface{} {
	return func(c *gin.Context) {
		carrier := make(propagation.HeaderCarrier)
		traceparent := c.GetHeader("traceparent")
		if traceparent != "" {
			carrier.Set("traceparent", traceparent)
		}

		ctx, _ := h.TraceSvc.Extract(carrier)
		ctx, span := h.TraceSvc.Start(ctx, fmt.Sprintf("http.request-%s", c.Request.URL.Path),
			"http.method", c.Request.Method,
			"http.url", c.Request.URL.Path,
			"net.peer.ip", c.ClientIP())

		spanCtx := otelTrace.SpanFromContext(ctx).SpanContext()
		traceId := ""
		if spanCtx.HasTraceID() {
			traceId = spanCtx.TraceID().String()
			spanId := spanCtx.SpanID().String()
			// 添加 traceId 到 span 属性
			span.SetAttribute("trace.id", traceId)

			// 将 traceId 注入到响应头
			c.Header("X-Trace-Id", traceId)
			c.Header("traceparent", "00-"+traceId+"-"+spanId+"-01")
		}

		defer span.End()

		// 将 traceId 存储到 context.Context 中（不是 gin.Context）
		ctx = context.WithValue(ctx, trace.TraceIdKey, traceId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttribute("http.status_code", c.Writer.Status())
	}
}
