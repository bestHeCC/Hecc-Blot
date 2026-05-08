package cache

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"core/config/cache"

	"github.com/stretchr/testify/assert"
)

var (
	total = len(MockLocalData)
	fail  = 0
)

var mockConf = cache.Local{
	ClearInterval: 1,
}

func TestLocalSet(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf)

	for _, k := range MockLocalData {
		err := mockCache.Set(k.key, k.val, k.expire)
		if err != nil {
			fail++
		}
	}

	assert.Equalf(t, 0, fail, "缓存未保存成功")
}

func TestLocalGet(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf)

	for _, k := range MockLocalData {
		_ = mockCache.Set(k.key, k.val, k.expire)

		val, err := mockCache.Get(k.key)
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
}

func TestLocalExists(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf)
	for _, k := range MockLocalData {
		_ = mockCache.Set(k.key, k.val, k.expire)
	}

	// 初始化随机种子
	rand.NewSource(time.Now().UnixNano())

	// 随机选择一个元素
	randomIndex := rand.Intn(total)
	randData := MockLocalData[randomIndex]
	t.Log(randData)

	ok, err := mockCache.Exists(randData.key)

	assert.Equal(t, nil, err)
	assert.Equalf(t, true, ok, "数据不存在")
}

func TestLocalDel(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf)

	for _, k := range MockLocalData {
		_ = mockCache.Set(k.key, k.val, k.expire)
	}

	// 初始化随机种子
	rand.NewSource(time.Now().UnixNano())

	// 随机选择一个元素
	randomIndex := rand.Intn(total)
	randData := MockLocalData[randomIndex]
	t.Log(randData)

	err := mockCache.Del(randData.key)
	assert.Equal(t, nil, err)

	v, err := mockCache.Get(randData.key)
	assert.Equal(t, nil, err)
	assert.Equalf(t, nil, v, "数据未删除")
}

func TestLocalClear(t *testing.T) {
	mockCache := newLocalCacheSvc(&mockConf)

	for _, k := range MockLocalData {
		_ = mockCache.Set(k.key, k.val, k.expire)

		val, _ := mockCache.Get(k.key)
		t.Log(k.key, val)
	}

	t.Log("============ sleep 16s start ============")
	time.Sleep(time.Second * 16)
	t.Log("============ sleep 16s end ============")

	for _, k := range MockLocalData {
		val, _ := mockCache.Get(k.key)
		if val != nil {
			fail++
		}
		t.Log(k.key, val)
	}

	assert.Equalf(t, 0, fail, "clear fail")
}
