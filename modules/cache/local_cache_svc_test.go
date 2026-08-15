package cache

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/bestHeCC/hecc-core/entity/config/cache"
	"github.com/bestHeCC/hecc-trace"

	"github.com/stretchr/testify/assert"
)

var (
	total = len(mockLocalData)
	fail  = 0
)

var mockConf = cache.Local{
	ClearInterval: 1,
}

var mockLocalData = []struct {
	key    string
	val    interface{}
	expire time.Duration
}{
	{"hcc1", 1, time.Second * 10},
	{"hcc2", "中文字符串", time.Second * 11},
	{"hcc3", "hcc is the best", time.Second * 12},
	{"hcc4", true, time.Second * 13},
	{"hcc5", false, time.Second * 14},
	{"hcc6", map[string]interface{}{"a": 1, "b": "中文", "c": map[string]string{"aa": "hh"}}, time.Second * 15},
}

func TestLocalCacheSvc(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf, nil)

	ctx := context.Background()
	t.Run("set", func(t *testing.T) {
		for _, k := range mockLocalData {
			err := mockCache.Set(ctx, k.key, k.val, k.expire)
			if err != nil {
				fail++
			}
		}

		assert.Equalf(t, 0, fail, "缓存未保存成功")
	})

	t.Run("get", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)

			val, err := mockCache.Get(ctx, k.key)
			if err != nil {
				fail++
			}

			// 取出map类型时，用反射比较
			if reflect.TypeOf(val).Kind() == reflect.Map {
				if !reflect.DeepEqual(val, k.val) {
					fail++
				}
			} else {
				if val != k.val {
					fail++
				}
			}
		}

		assert.Equalf(t, 0, fail, "缓存未获取成功")
	})

	t.Run("exists", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := mockLocalData[randomIndex]
		t.Log(randData)

		ok, err := mockCache.Exists(ctx, randData.key)

		assert.Equal(t, nil, err)
		assert.Equalf(t, true, ok, "数据不存在")
	})

	t.Run("delWithLock", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := mockLocalData[randomIndex]
		t.Log(randData)

		err := mockCache.Del(ctx, randData.key)
		assert.Equal(t, nil, err)

		v, err := mockCache.Get(ctx, randData.key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, nil, v, "数据未删除")
	})

	t.Run("clear", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)

			val, _ := mockCache.Get(ctx, k.key)
			t.Log(k.key, val)
		}

		t.Log("============ sleep 16s start ============")
		time.Sleep(time.Second * 16)
		t.Log("============ sleep 16s end ============")

		for _, k := range mockLocalData {
			val, _ := mockCache.Get(ctx, k.key)
			if val != nil {
				fail++
			}
			t.Log(k.key, val)
		}

		assert.Equalf(t, 0, fail, "clear fail")
	})
}

func TestLocalCacheSvcWithTrace(t *testing.T) {
	traceSvc, traceClearUp, err := trace.NewTraceSvc(&mockTraceConf)
	defer traceClearUp()

	assert.Equal(t, nil, err)

	mockCache := newLocalCacheSvc(&mockConf, traceSvc)
	ctx := context.Background()
	t.Run("set", func(t *testing.T) {
		for _, k := range mockLocalData {
			err := mockCache.Set(ctx, k.key, k.val, k.expire)
			if err != nil {
				fail++
			}
		}

		assert.Equalf(t, 0, fail, "缓存未保存成功")
	})

	t.Run("get", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)

			val, err := mockCache.Get(ctx, k.key)
			if err != nil {
				fail++
			}

			// 取出map类型时，用反射比较
			if reflect.TypeOf(val).Kind() == reflect.Map {
				if !reflect.DeepEqual(val, k.val) {
					fail++
				}
			} else {
				if val != k.val {
					fail++
				}
			}
		}

		assert.Equalf(t, 0, fail, "缓存未获取成功")
	})

	t.Run("exists", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := mockLocalData[randomIndex]
		t.Log(randData)

		ok, err := mockCache.Exists(ctx, randData.key)

		assert.Equal(t, nil, err)
		assert.Equalf(t, true, ok, "数据不存在")
	})

	t.Run("delWithLock", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := mockLocalData[randomIndex]
		t.Log(randData)

		err = mockCache.Del(ctx, randData.key)
		assert.Equal(t, nil, err)

		v, err := mockCache.Get(ctx, randData.key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, nil, v, "数据未删除")
	})

	t.Run("clear", func(t *testing.T) {
		for _, k := range mockLocalData {
			_ = mockCache.Set(ctx, k.key, k.val, k.expire)

			val, _ := mockCache.Get(ctx, k.key)
			t.Log(k.key, val)
		}

		t.Log("============ sleep 16s start ============")
		time.Sleep(time.Second * 16)
		t.Log("============ sleep 16s end ============")

		for _, k := range mockLocalData {
			val, _ := mockCache.Get(ctx, k.key)
			if val != nil {
				fail++
			}
			t.Log(k.key, val)
		}

		assert.Equalf(t, 0, fail, "clear fail")
	})
}
