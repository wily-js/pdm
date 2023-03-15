package middle

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

type EditLock struct {
	editCache *cache.Cache
}

const (
	NoLock = 0 // 无人持有该锁
)

func NewEditLock() *EditLock {
	res := &EditLock{}
	res.editCache = cache.New(5*time.Hour, 7*time.Hour)
	return res
}

// Lock 加锁
func (c *EditLock) Lock(userId int, projectId int, id int, docType string) {
	key := fmt.Sprintf("type:%s-%d-%d", docType, id, projectId)
	c.editCache.Set(key, userId, cache.DefaultExpiration)
}

// Query 查询锁 0 - 无人持有该锁 其他 - 持有该锁的用户ID
func (c *EditLock) Query(projectId int, id int, docType string) interface{} {
	key := fmt.Sprintf("type:%s-%d-%d", docType, id, projectId)
	v, found := c.editCache.Get(key)
	if found != false {
		return v
	} else {
		return 0
	}
}

// Unlock 解锁
func (c *EditLock) Unlock(projectId int, id int, docType string) {
	key := fmt.Sprintf("type:%s-%d-%d", docType, id, projectId)
	c.editCache.Delete(key)
}
