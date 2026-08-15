package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bestHeCC/hecc-api"
	"github.com/bestHeCC/hecc-cache"
	iCoreApi "github.com/bestHeCC/hecc-core/contract/api"
	iCoreCache "github.com/bestHeCC/hecc-core/contract/cache"
	iCoreDb "github.com/bestHeCC/hecc-core/contract/db"
	iCoreError "github.com/bestHeCC/hecc-core/contract/error"
	iCoreLog "github.com/bestHeCC/hecc-core/contract/log"
	iCoreSse "github.com/bestHeCC/hecc-core/contract/sse"
	iCoreTrace "github.com/bestHeCC/hecc-core/contract/trace"
	entityApi "github.com/bestHeCC/hecc-core/entity/api"
	coreConfig "github.com/bestHeCC/hecc-core/entity/config"
	dbEnum "github.com/bestHeCC/hecc-core/enum/db"
	"github.com/bestHeCC/hecc-core/enum/response"
	"github.com/bestHeCC/hecc-core/util"
	"github.com/bestHeCC/hecc-db"
	errorSvc "github.com/bestHeCC/hecc-error"
	"github.com/bestHeCC/hecc-log"
	"github.com/bestHeCC/hecc-sse"
	"github.com/bestHeCC/hecc-trace"

	"github.com/bestHeCC/hecc-ioc"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/plugin/soft_delete"
)

// ===== 1. 启动入口 =====
// 演示：框架初始化全流程 — 配置 → 日志 → 追踪 → 数据库 → 缓存 → IOC → 路由 → 启动
// 详见：docs/quick_start.md

func main() {
	config := initConf("config.yaml")

	logSvc := must(log.NewLogger(&config.Log))
	traceSvc, traceClearUp := must2(trace.NewTraceSvc(&config.Trace))
	dbFactory, dbClearUp := must2(db.NewDbFactory(&config.Db, logSvc))

	cacheFactory := cache.NewCacheFactory(&config.Cache, traceSvc)
	responseSvc := api.NewResponseSvc()

	// defer 注册退出清理（LIFO 顺序执行）
	defer func() {
		dbClearUp()
		traceClearUp()
		if cacheFactory.Redis() != nil {
			_ = cacheFactory.Redis().Close()
		}
	}()

	// 注册到 IOC 容器（顺序无关，但必须在路由注册之前）
	container := ioc.New()

	container.Set(new(iCoreDb.IDbFactory), dbFactory)
	container.Set(new(iCoreLog.ILog), logSvc)
	container.Set(new(iCoreCache.ICacheFactory), cacheFactory)
	container.Set(new(iCoreApi.IResponse), responseSvc)
	container.Set(new(iCoreTrace.ITrace), traceSvc)

	apiHandle := api.NewApiSvc(&config.Server, responseSvc, traceSvc, container)
	registerRoutes(apiHandle)

	sseHandle := sse.NewSseSvc(apiHandle.Engine(), container)
	registerSseRoutes(sseHandle)

	apiHandle.Listen()
}

// must 单返回值错误处理：出错直接 panic
func must[T any](val T, err error) T {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val
}

// must2 双返回值错误处理：出错直接 panic
func must2[T, U any](val T, cleanup U, err error) (T, U) {
	if err != nil {
		panic(fmt.Errorf("初始化失败: %w", err))
	}
	return val, cleanup
}

// ===== 2. 配置加载 =====
// 演示：使用 viper 读取 config.yaml，反序列化为 config.Config 结构体
// 详见：docs/config.md

func initConf(configPath string) *coreConfig.Config {
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %w", err))
	}
	var conf coreConfig.Config
	if err := v.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("解析配置文件失败: %w", err))
	}
	return &conf
}

// ===== 3. Model 定义 =====
// 演示：实现 IDbModel 接口（GetID），定义表名（TableName），支持多 Model
// 详见：docs/database.md

// AccountModel 账户模型
type AccountModel struct {
	ID          int                   `json:"id" gorm:"primaryKey"`
	AccountName string                `json:"account_name"`
	Password    string                `json:"password"`
	Email       string                `json:"email"`
	Balance     float64               `json:"balance"`
	CreatedAt   int64                 `json:"created_at"`
	UpdatedAt   int64                 `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"deleted_at"`
}

func (a AccountModel) TableName() string { return "account" }

func (a AccountModel) GetID() int { return a.ID }

// OrderModel 订单模型 — 演示多 Model 场景
type OrderModel struct {
	ID        int     `json:"id" gorm:"primaryKey"`
	AccountID int     `json:"account_id"`
	Product   string  `json:"product"`
	Amount    float64 `json:"amount"`
	CreatedAt int64   `json:"created_at"`
}

func (o OrderModel) TableName() string { return "order" }

func (o OrderModel) GetID() int { return o.ID }

// ===== 4. 请求参数与校验 =====
// 演示：binding tag 自动校验（required/min/max/email）、自定义错误信息 GetMessages()
// 详见：docs/routes_middleware.md

// AddAccountRequest 新增账户 — 展示多种校验规则
type AddAccountRequest struct {
	AccountName string `json:"account_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	Email       string `json:"email" binding:"required,email"`
	Age         int    `json:"age" binding:"min=1,max=150"`
}

func (a AddAccountRequest) GetMessages() entityApi.Messages {
	return entityApi.Messages{
		"AccountName.required": "用户名不能为空",
		"Password.required":    "密码不能为空",
		"Password.min":         "密码长度不能少于6位",
		"Email.required":       "邮箱不能为空",
		"Email.email":          "邮箱格式不正确",
		"Age.min":              "年龄不能小于1",
		"Age.max":              "年龄不能大于150",
	}
}

// ===== 5. 中间件 =====
// 演示：定义 Token 校验中间件，中间件中通过 inject tag 注入依赖
// 详见：docs/routes_middleware.md

// TokenMiddleware Token 鉴权中间件
type TokenMiddleware struct {
	ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() any {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, fmt.Errorf("缺少 Authorization 头")))
			c.Abort()
			return
		}
		// 实际项目可在此解析 JWT、查询用户信息等
		c.Set("token", token)
		c.Next()
	}
}

// SseAcceptMiddleware SSE Accept 头校验中间件
// 演示：策略性校验（如 Accept 头）通过中间件实现，而非框架内置
type SseAcceptMiddleware struct{}

func (m SseAcceptMiddleware) Middleware() any {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept"), "text/event-stream") {
			c.String(http.StatusNotAcceptable, "Accept: text/event-stream required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ===== 6. 数据库 CRUD =====
// 演示：Add / Take / Find / Select / Save / Remove / Order / Count + 事务 Begin/Commit/Rollback
// 详见：docs/database.md

// AddAccountApi 新增账户 + 事务演示
type AddAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
	LogSvc    iCoreLog.ILog      `inject:""`
	AddAccountRequest
}

func (a AddAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	account := AccountModel{
		AccountName: a.AccountName,
		Password:    a.Password,
		Email:       a.Email,
	}

	db := a.DbFactory.Build(ctx)

	// 开启事务
	tx := db.Begin()
	if err := tx.Add(&account); err != nil {
		tx.Rollback()
		return nil, errorSvc.NewError(response.Fail, err)
	}
	// 同时写入关联订单
	order := OrderModel{AccountID: account.ID, Product: "新用户礼包", Amount: 0}
	if err := tx.Add(&order); err != nil {
		tx.Rollback()
		return nil, errorSvc.NewError(response.Fail, err)
	}
	if err := tx.Commit(); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	a.LogSvc.Info(ctx, "account created", "id", account.ID)
	return account, nil
}

// TakeAccountApi 查询单条记录
type TakeAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a TakeAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	var account AccountModel
	if err := db.Where("id = ?", 1).Take(&account); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return account, nil
}

// FindAccountApi 查询多条记录（条件筛选 + 排序 + 字段选择）
type FindAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a FindAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	var list []AccountModel
	if err := db.
		Select("id, account_name, email").
		Where("id >= ?", 1).
		Order("id DESC").
		Find(&list); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return list, nil
}

// CountAccountApi 统计记录数
type CountAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a CountAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	count, err := db.Query(AccountModel{}).Where("id >= ?", 1).Count()
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return count, nil
}

// UpdateAccountApi 更新记录
type UpdateAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a UpdateAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	updateData := AccountModel{AccountName: "updated_name", Email: "new@example.com"}
	if err := db.Where("id = ?", 1).Save(&updateData); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return updateData, nil
}

// DeleteAccountApi 删除记录
type DeleteAccountApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a DeleteAccountApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx)
	if err := db.Where("id = ?", 1).Remove(&AccountModel{}); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}
	return nil, nil
}

// ===== 7. 多数据库切换 =====
// 演示：SetDefault() 切换默认库、Build(ctx, dbEnum.xxx) 运行时指定数据库
// 详见：docs/database.md

// DbSwitchApi 多数据库切换 — 展示同时操作 MySQL 和 PostgreSQL
type DbSwitchApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
}

func (a DbSwitchApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 方式一：使用默认数据库（通常是 MySQL）
	mysqlDB := a.DbFactory.Build(ctx)

	// 方式二：运行时指定数据库类型
	pgDB := a.DbFactory.Build(ctx, dbEnum.Postgres)

	// 分别从两个数据库查询
	var mysqlAccounts []AccountModel
	if err := mysqlDB.Where("id >= ?", 1).Find(&mysqlAccounts); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	var pgAccounts []AccountModel
	if err := pgDB.Where("id >= ?", 1).Find(&pgAccounts); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 还可以运行时切换默认库
	// a.DbFactory.SetDefault(dbEnum.Postgres)

	return map[string]interface{}{
		"mysql": mysqlAccounts,
		"pg":    pgAccounts,
	}, nil
}

// ===== 8. 缓存操作 =====
// 演示：本地缓存 + Redis 缓存的 Get/Set/Del/Exists、Redis Hash 操作、缓存穿透回写
// 详见：docs/cache.md

// CacheBasicApi 缓存基础操作
type CacheBasicApi struct {
	CacheFactory iCoreCache.ICacheFactory `inject:""`
}

func (a CacheBasicApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 本地缓存 — Set / Get / Exists / Del
	_ = a.CacheFactory.Local().Set(ctx, "local:key", "hello", 10*time.Minute)

	if ok, _ := a.CacheFactory.Local().Exists(ctx, "local:key"); ok {
		val, _ := a.CacheFactory.Local().Get(ctx, "local:key")
		_ = a.CacheFactory.Local().Del(ctx, "local:key")
		_ = val
	}

	// Redis 缓存 — Set / Get / Del
	_ = a.CacheFactory.Redis().Set(ctx, "redis:key", "world", time.Hour)
	val, _ := a.CacheFactory.Redis().Get(ctx, "redis:key")
	_ = a.CacheFactory.Redis().Del(ctx, "redis:key")

	return val, nil
}

// CacheHashApi Redis Hash 操作
type CacheHashApi struct {
	CacheFactory iCoreCache.ICacheFactory `inject:""`
}

func (a CacheHashApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// HSet — 同时设置多个 field
	err := a.CacheFactory.Redis().HSet(ctx, "user:1", "name", "john", "email", "john@test.com")
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// HGet — 获取单个 field
	name, _ := a.CacheFactory.Redis().HGet(ctx, "user:1", "name")

	// HDel — 删除指定 field
	_ = a.CacheFactory.Redis().HDel(ctx, "user:1", "email")

	return name, nil
}

// CacheReadThroughApi 缓存读穿透 — 先查缓存，未命中则查 DB 并回写缓存
type CacheReadThroughApi struct {
	CacheFactory iCoreCache.ICacheFactory `inject:""`
	DbFactory    iCoreDb.IDbFactory       `inject:""`
}

func (a CacheReadThroughApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	cacheKey := "account:1"

	// 1. 先从本地缓存读取
	if cached, _ := a.CacheFactory.Local().Get(ctx, cacheKey); cached != nil {
		return cached, nil
	}

	// 2. 缓存未命中，查数据库
	db := a.DbFactory.Build(ctx)
	var account AccountModel
	if err := db.Where("id = ?", 1).Take(&account); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// 3. 回写缓存（本地 + Redis 双写）
	_ = a.CacheFactory.Local().Set(ctx, cacheKey, account, 10*time.Minute)
	_ = a.CacheFactory.Redis().Set(ctx, cacheKey, account, 10*time.Minute)

	return account, nil
}

// ===== 9. 链路追踪 =====
// 演示：FromContext / SetAttribute / RecordError / Start 子 Span / defer span.End()
// 详见：docs/trace.md

// TraceDemoApi 链路追踪示例
type TraceDemoApi struct {
	TraceSvc iCoreTrace.ITrace `inject:""`
	LogSvc   iCoreLog.ILog     `inject:""`
}

func (a TraceDemoApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	// 1. 从 Context 获取当前请求的 Span（由 HttpTraceMiddleware 自动创建）
	currentSpan := a.TraceSvc.FromContext(ctx)

	// 2. 为当前 Span 添加自定义属性
	currentSpan.SetAttribute("business.type", "trace_demo")
	currentSpan.SetAttribute("user.id", 12345)

	// 3. 开启子 Span 追踪数据库操作
	subCtx, subSpan := a.TraceSvc.Start(ctx, "db-slow-query",
		"db.table", "account",
		"db.operation", "find",
	)
	defer subSpan.End()

	// 模拟耗时操作
	time.Sleep(10 * time.Millisecond)

	// 4. 模拟出错时记录错误到 Span
	if false { // 实际业务中将条件替换为 err != nil
		subSpan.RecordError(fmt.Errorf("模拟数据库错误"))
	}

	a.LogSvc.Info(subCtx, "trace demo completed", "span", subSpan.Name())
	return "trace demo ok", nil
}

// ===== 10. 分页 =====
// 演示：Offset/limit 分页（NewPage）+ 游标分页（NewCursor）
// 详见：docs/paginator.md

// PageRequest offset 分页请求参数
type PageRequest struct {
	Page     int `json:"page" binding:"min=1"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

// PageListApi offset/limit 分页示例
type PageListApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
	PageRequest
}

func (a PageListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	opts := util.PageOpts{Page: a.Page, PageSize: a.PageSize}
	db := a.DbFactory.Build(ctx).Query(AccountModel{})

	total, err := db.Count()
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	var list []AccountModel
	offset := (opts.Page - 1) * opts.PageSize
	if err = db.Order("id DESC").Limit(opts.PageSize).Offset(offset).Find(&list); err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// NewPage 自动处理 nil → []、默认 page/pageSize
	return util.NewPage(list, total, opts), nil
}

// CursorRequest 游标分页请求参数
type CursorRequest struct {
	Cursor   int `json:"cursor"`
	PageSize int `json:"pageSize" binding:"min=1,max=100"`
}

// CursorListApi 游标分页示例
type CursorListApi struct {
	DbFactory iCoreDb.IDbFactory `inject:""`
	CursorRequest
}

func (a CursorListApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
	db := a.DbFactory.Build(ctx).Query(AccountModel{})

	// 多查一条用于判断 hasMore
	var list []AccountModel
	err := db.Where("id > ?", a.Cursor).Order("id ASC").Limit(a.PageSize + 1).Find(&list)
	if err != nil {
		return nil, errorSvc.NewError(response.Fail, err)
	}

	// NewCursor 自动判断 hasMore 并截断多余数据
	// func(item *AccountModel) any 提取游标值（这里用 ID 作为游标）
	return util.NewCursor(list, a.PageSize, func(item *AccountModel) any {
		return item.ID
	}), nil
}

// ===== 11. SSE 推送 =====
// 演示：ISse 接口 + Writer 写入抽象（心跳、Flusher 断言由框架处理）
// 详见：docs/sse.md

// ExampleSse SSE 实时推送示例
type ExampleSse struct {
	LogSvc iCoreLog.ILog `inject:""`
}

func (e ExampleSse) Serve(ctx context.Context, w iCoreSse.Writer) error {
	e.LogSvc.Info(ctx, "sse connection established")

	// 业务推送：每秒推送服务器时间
	business := time.NewTicker(1 * time.Second)
	defer business.Stop()

	for {
		select {
		case <-ctx.Done():
			// 客户端断开或心跳写入失败
			e.LogSvc.Info(ctx, "sse client disconnected")
			return nil
		case <-business.C:
			msg := fmt.Sprintf("当前服务器时间：%s", time.Now().Format(time.RFC3339))
			if err := w.Send("", "", msg); err != nil {
				return err
			}
		}
	}
}

// ==============================
// 路由注册（集中管理）
// ==============================

func registerRoutes(apiHandle iCoreApi.IApiHandle) {
	// API 路由分组，Token 鉴权中间件仅作用于该分组（SSE 不受影响）
	apiGroup := apiHandle.Group("", &TokenMiddleware{})

	{
		// — Section 4: 参数校验 —
		apiGroup.Post("account/add", &AddAccountApi{})

		// — Section 6: 数据库 CRUD —
		apiGroup.Get("account/take", &TakeAccountApi{})
		apiGroup.Get("account/find", &FindAccountApi{})
		apiGroup.Get("account/count", &CountAccountApi{})
		apiGroup.Post("account/update", &UpdateAccountApi{})
		apiGroup.Post("account/delete", &DeleteAccountApi{})

		// — Section 7: 多数据库切换 —
		apiGroup.Get("account/db-switch", &DbSwitchApi{})

		// — Section 8: 缓存操作 —
		apiGroup.Get("cache/basic", &CacheBasicApi{})
		apiGroup.Get("cache/hash", &CacheHashApi{})
		apiGroup.Get("cache/read-through", &CacheReadThroughApi{})

		// — Section 9: 链路追踪 —
		apiGroup.Get("trace/demo", &TraceDemoApi{})

		// — Section 10: 分页 —
		apiGroup.Post("account/page", &PageListApi{})
		apiGroup.Post("account/cursor", &CursorListApi{})
	}
}

func registerSseRoutes(sseHandle iCoreSse.ISseHandle) {
	// — Section 11: SSE 推送 —
	// 通过中间件做 Accept 校验（策略性校验不内置在框架）
	sseGroup := sseHandle.Group("", &SseAcceptMiddleware{})

	// GET 方式：EventSource 标准用法
	sseGroup.Get("events/time", &ExampleSse{})

	// POST 方式：适用于 fetch + ReadableStream（可携带请求体）
	sseGroup.Post("events/time", &ExampleSse{})
}
