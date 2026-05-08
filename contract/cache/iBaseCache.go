package cache

import (
	"time"
)

type IBaseCache interface {
	// Set 将value写入缓存
	Set(key string, val interface{}, expire time.Duration) error
	// Get 根据key值获取value
	Get(key string) (interface{}, error)
	// Del 删除key值
	Del(key string) error
	// Exists 判断key是否存在
	Exists(key string) (bool, error)
}
