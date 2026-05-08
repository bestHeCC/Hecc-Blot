package cache

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"core/config/cache"

	"github.com/stretchr/testify/assert"
)

var mockFactoryConf = cache.Config{
	Local: cache.Local{
		ClearInterval: 1,
	},
	Redis: cache.Redis{
		Addr:     "127.0.0.1:6379",
		Password: "",
		PoolSize: 10,
		DB:       0,
	},
}

var MockLocalData = []struct {
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

func TestFactory4Local(t *testing.T) {
	factory := NewCacheFactory(&mockFactoryConf)
	local := factory.Local()

	t.Run("local cache set", func(t *testing.T) {
		for _, k := range MockLocalData {
			err := local.Set(k.key, k.val, k.expire)
			if err != nil {
				fail++
			}
		}

		assert.Equalf(t, 0, fail, "缓存未保存成功")
	})

	t.Run("local cache get", func(t *testing.T) {
		for _, k := range MockLocalData {
			_ = local.Set(k.key, k.val, k.expire)

			val, err := local.Get(k.key)
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

	t.Run("local cache exists", func(t *testing.T) {
		for _, k := range MockLocalData {
			_ = local.Set(k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := MockLocalData[randomIndex]
		t.Log(randData)

		ok, err := local.Exists(randData.key)

		assert.Equal(t, nil, err)
		assert.Equalf(t, true, ok, "数据不存在")
	})

	t.Run("local cache del", func(t *testing.T) {
		for _, k := range MockLocalData {
			_ = local.Set(k.key, k.val, k.expire)
		}

		// 初始化随机种子
		rand.NewSource(time.Now().UnixNano())

		// 随机选择一个元素
		randomIndex := rand.Intn(total)
		randData := MockLocalData[randomIndex]
		t.Log(randData)

		err := local.Del(randData.key)
		assert.Equal(t, nil, err)

		v, err := local.Get(randData.key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, nil, v, "数据未删除")
	})

	t.Run("local cache clear", func(t *testing.T) {
		for _, k := range MockLocalData {
			_ = local.Set(k.key, k.val, k.expire)

			val, _ := local.Get(k.key)
			t.Log(k.key, val)
		}

		t.Log("============ sleep 16s start ============")
		time.Sleep(time.Second * 16)
		t.Log("============ sleep 16s end ============")

		for _, k := range MockLocalData {
			val, _ := local.Get(k.key)
			if val != nil {
				fail++
			}
			t.Log(k.key, val)
		}

		assert.Equalf(t, 0, fail, "clear fail")
	})
}

func TestFactory4Redis(t *testing.T) {
	factory := NewCacheFactory(&mockFactoryConf)
	redis := factory.Redis()

	t.Run("redis set", func(t *testing.T) {
		key := "hcc-set"

		err := redis.Set(key, "1", 0)
		assert.Equal(t, nil, err)
	})

	t.Run("redis get", func(t *testing.T) {
		key := "hcc-get"
		val := "1"
		_ = redis.Set(key, val, 0)

		v, err := redis.Get(key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "取值不一致")
	})

	t.Run("redis del", func(t *testing.T) {
		key := "hcc-del"
		val := "1"
		_ = redis.Set(key, val, 0)

		err := redis.Del(key)
		assert.Equal(t, nil, err)
	})

	t.Run("redis exists", func(t *testing.T) {
		key := "hcc-exists"
		val := "1"
		_ = redis.Set(key, val, 0)

		exists, err := redis.Exists(key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, true, exists, "值不存在")
	})

	t.Run("redis hset", func(t *testing.T) {
		key := "hcc-hset"
		field := "1"
		val := "1"
		err := redis.HSet(key, field, val)
		assert.Equal(t, nil, err)
	})

	t.Run("redis hget", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redis.HSet(key, field, val)
		assert.Equal(t, nil, err)

		v, err := redis.HGet(key, field)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "值不一致")
	})

	t.Run("redis hdel", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redis.HSet(key, field, val)
		assert.Equal(t, nil, err)

		err = redis.HDel(key, field)
		assert.Equal(t, nil, err)
	})

	t.Run("redis close", func(t *testing.T) {
		err := redis.Close()
		assert.Equal(t, nil, err)
	})
}
