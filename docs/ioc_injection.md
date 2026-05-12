# IOC 组件自动注入说明

## 概述

IOC（控制反转）是 Hecc-Go-Core 框架的核心组件，负责管理所有服务的生命周期和依赖注入。通过 IOC 容器，框架可以自动将依赖注入到需要的地方，无需手动创建和传递依赖。

---

## IOC 核心实现

### 1. 核心数据结构

```go
// 存储接口类型到实例的映射
var instanceValues = make(map[reflect.Type]map[string]reflect.Value)
```

- **外层 Map**: Key 为接口类型 (`reflect.Type`)，Value 为内层 Map
- **内层 Map**: Key 为实例名称（用于区分同接口的多个实现），Value 为实例的反射值

### 2. 注册方法

#### Set - 注册默认实例

```go
func Set(interfaceObj interface{}, instance interface{}) {
    SetWithName(interfaceObj, "", instance)
}
```

**参数说明**:
- `interfaceObj`: 接口类型（通常使用 `new(InterfaceType)` 获取）
- `instance`: 实现该接口的具体实例

**使用示例**:

```go
// 注册日志服务
logSvc := log.NewLogger(&config.Log)
ioc.Set(new(iCoreLog.ILog), logSvc)

// 注册数据库工厂
dbFactory, _, _ := db.NewDbFactory(&config.Db, logSvc)
ioc.Set(new(iCoreDb.IDbFactory), dbFactory)
```

#### SetWithName - 注册命名实例

```go
func SetWithName(interfaceObj interface{}, name string, instance interface{})
```

**使用场景**: 同一接口有多个实现时，通过名称区分

```go
// 注册两个不同的日志实现
ioc.SetWithName(new(iCoreLog.ILog), "local", localLog)
ioc.SetWithName(new(iCoreLog.ILog), "remote", remoteLog)
```

---

## 自动注入原理

### 1. 注入流程

```go
func Inject(instance interface{}) {
    instanceValue := reflect.ValueOf(instance)
    if instanceValue.Kind() != reflect.Ptr {
        panic("ioc: 注入实例必须是指针")
    }
    inject(instanceValue)
}
```

**注入步骤**:

1. **检查类型**: 确保传入的是指针类型
2. **获取元素**: 通过 `Elem()` 获取指针指向的实际值
3. **遍历字段**: 遍历结构体的所有字段
4. **查找 inject tag**: 检查字段是否有 `inject` 标签
5. **获取实例**: 根据字段类型和名称从容器中获取实例
6. **设置值**: 将实例赋值给字段

### 2. 核心注入逻辑

```go
func inject(instanceValue reflect.Value) {
    if instanceValue.Kind() == reflect.Ptr {
        instanceValue = instanceValue.Elem()
    }
    
    instanceType := instanceValue.Type()
    for j := 0; j < instanceType.NumField(); j++ {
        field := instanceValue.Type().Field(j)
        fieldValue := instanceValue.FieldByIndex(field.Index)
        
        // 处理匿名嵌套结构体
        if field.Anonymous {
            if field.Type.Kind() == reflect.Struct {
                inject(fieldValue)
            }
            continue
        }
        
        // 查找 inject tag
        name, ok := field.Tag.Lookup("inject")
        if !ok {
            return  // 没有 inject tag，停止注入（约定：注入字段必须放在前面）
        }
        
        // 如果字段是指针类型，先创建实例
        if fieldValue.Kind() == reflect.Ptr {
            value := reflect.New(field.Type.Elem())
            fieldValue.Set(value)
            fieldValue = fieldValue.Elem()
        }
        
        // 从容器获取实例并设置
        v := getValueWithName(field.Type, name)
        fieldValue.Set(v)
    }
}
```

---

## 使用方式

### 1. 在 API 中注入依赖

```go
type AddApi struct {
    // 通过 inject tag 标记需要注入的字段
    DbFactory iCoreDb.IDbFactory `inject:""`
    LogSvc    iCoreLog.ILog      `inject:""`
    
    // 请求参数（必须放在最后）
    AddRequest
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    // 直接使用注入的服务
    a.LogSvc.Info(ctx, "add account")
    err := a.DbFactory.Build(ctx).Add(&account)
    return nil, nil
}
```

### 2. 在中间件中注入依赖

```go
type TokenMiddleware struct {
    ResponseSvc iCoreApi.IResponse `inject:""`
}

func (t TokenMiddleware) Middleware() interface{} {
    return func(c *gin.Context) {
        // 使用注入的响应服务
        t.ResponseSvc.Regular(c, nil, errorSvc.NewError(response.TokenInvalid, err))
        c.Abort()
    }
}
```

### 3. 使用命名注入

```go
type CustomApi struct {
    // 指定注入名为 "custom" 的日志实例
    LogSvc iCoreLog.ILog `inject:"custom"`
}
```

---

## 替换框架组件并保持自动注入

### 场景说明

当对框架默认组件不满意时，可以替换为自定义实现，同时保持自动注入功能。

### 替换步骤

#### 1. 实现接口

```go
// 自定义日志实现
type MyLogSvc struct{}

func (m MyLogSvc) Debug(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyLogSvc) Error(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyLogSvc) Info(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}

func (m MyLogSvc) Warn(ctx context.Context, msg string, fields ...interface{}) {
    // 自定义实现
}
```

#### 2. 注册到 IOC 容器

```go
func main() {
    // 创建自定义实例
    myLog := &MyLogSvc{}
    
    // 注册到 IOC 容器（覆盖默认实现）
    ioc.Set(new(iCoreLog.ILog), myLog)
    
    // 创建 API 处理器
    apiHandle := api.NewApiSvc(&config.Server, responseSvc)
    
    // 注册 API（自动注入时会使用自定义实现）
    apiHandle.Post("account/add", &AddApi{})
    
    apiHandle.Listen()
}
```

#### 3. 验证注入

```go
type AddApi struct {
    LogSvc iCoreLog.ILog `inject:""`  // 会自动注入 MyLogSvc
}

func (a AddApi) Call(ctx *gin.Context) (interface{}, iCoreError.IError) {
    // 使用的是自定义的 MyLogSvc
    a.LogSvc.Info(ctx, "using custom log service")
    return nil, nil
}
```

---

## 注入规则

### 1. 字段顺序

**重要**: 注入字段必须放在结构体的**最前面**，请求参数放在最后面。

```go
// 正确
type AddApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`  // 注入字段在前
    LogSvc    iCoreLog.ILog      `inject:""`
    AddRequest                          // 请求参数在后
}

// 错误 - 注入字段和请求参数混合
type AddApi struct {
    DbFactory iCoreDb.IDbFactory `inject:""`
    Name      string             // 请求参数
    LogSvc    iCoreLog.ILog      `inject:""`  // 不会被注入
}
```

### 2. 匿名结构体处理

IOC 支持匿名嵌套结构体的注入：

```go
type BaseApi struct {
    LogSvc iCoreLog.ILog `inject:""`
}

type AddApi struct {
    BaseApi               // 匿名嵌套，会自动注入
    DbFactory iCoreDb.IDbFactory `inject:""`
}
```

### 3. 指针类型支持

注入字段可以是指针类型：

```go
type MyApi struct {
    LogSvc *MyLogSvc `inject:""`  // 指针类型也支持
}
```

---

## 单测示例

查看 `service/ioc/iocSvc_test.go` 了解 IOC 的测试方式：

```go
func TestIocSvc(t *testing.T) {
    // 注册实例
    Set(new(iInterface), derive{})
    SetWithName(new(iInterface), "custom", derive{})
    
    // 测试默认注入
    t.Run("default", func(t *testing.T) {
        var d1 defaultTest
        Inject(&d1)
        assert.Equal(t, "set test", d1.One.Test())
    })
    
    // 测试命名注入
    t.Run("custom", func(t *testing.T) {
        var d2 customTest
        Inject(&d2)
        assert.Equal(t, "set test", d2.One.Test())
    })
}
```

---

## IOC 工作流程图

```
┌─────────────────────────────────────────────────────────────┐
│                        IOC 工作流程                         │
├─────────────────────────────────────────────────────────────┤
│                                                            │
│  1. 注册阶段                                                │
│     ┌─────────────┐    Set()    ┌──────────────────┐       │
│     │ 创建实例     │ ──────────→ │ 存入 IOC 容器     │       │
│     │ logSvc      │             │ interfaceType    │       │
│     │ dbFactory   │             │ → instance       │       │
│     └─────────────┘             └──────────────────┘       │
│                                                            │
│  2. 注入阶段                                                │
│     ┌─────────────┐   Inject()  ┌──────────────────┐       │
│     │ API 结构体   │ ──────────→ │ 遍历字段         │       │
│     │ AddApi      │             │ 查找 inject tag  │       │
│     └─────────────┘             │ 从容器获取实例    │       │
│                                 │ 设置字段值        │       │
│                                 └──────────────────┘       │
│                                                            │
│  3. 使用阶段                                                │
│     ┌─────────────┐             ┌──────────────────┐       │
│     │ a.LogSvc    │ ←───────── │ 自动注入完成      │       │
│     │ .Info(...)  │             │ 可直接使用       │       │
│     └─────────────┘             └──────────────────┘       │
│                                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 总结

IOC 组件实现了：

1. **依赖注入**: 通过 `inject` tag 自动注入依赖
2. **接口解耦**: 依赖接口而非具体实现
3. **组件替换**: 支持自定义实现替换默认组件
4. **生命周期管理**: 统一管理所有服务的生命周期

核心优势：
- **降低耦合**: 模块之间通过接口交互，不依赖具体实现
- **提高可测试性**: 可以方便地注入 Mock 实现进行测试
- **增强扩展性**: 新增功能只需实现接口并注册到容器
