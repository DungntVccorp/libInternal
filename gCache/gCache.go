package gCache

import (
	"context"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
)

type GCache struct {
	rdp        redis.UniversalClient
	mc         *memcache.Client
	ctx        context.Context
	cache_type int // 0 redis cache , 1 memcache
}

func NewRedisCache(sentinelAddrs []string, userName string, password string, database int, masterName string, clientName string) (*GCache, bool) {

	rs_config := &redis.UniversalOptions{
		Addrs:        sentinelAddrs,
		Username:     userName,
		Password:     password,
		DB:           database,
		MasterName:   masterName,
		ClientName:   clientName,
		ReadTimeout:  time.Second * 2,
		WriteTimeout: time.Second * 2,
		DialTimeout:  time.Second * 5,
		MaxRetries:   5,
		PoolSize:     10,
	}
	ctx := context.Background()

	rdp := redis.NewUniversalClient(rs_config)
	status := rdp.Ping(context.Background())
	if status.Err() != nil {
		fmt.Println(status.Err())
		return nil, false
	}
	return &GCache{rdp: rdp, ctx: ctx, cache_type: 0}, true
}
func NewMemcache(Addrs []string) (*GCache, bool) {
	p := GCache{
		mc:         memcache.New(Addrs...),
		cache_type: 1,
	}
	if err := p.mc.Ping(); err != nil {
		fmt.Println(err.Error())
		return nil, false
	}
	return &p, true

}

func (p *GCache) Store(key string, value string, ttl int) bool {
	if p.cache_type == 0 {
		err := p.rdp.Set(p.ctx, key, value, time.Duration(ttl)).Err()
		if err != nil {
			return false
		}
		return true
	} else if p.cache_type == 1 {
		err := p.mc.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: int32(ttl)})
		if err != nil {
			return false
		}
		return true

	}
	return false
}
func (p *GCache) Get(key string) (bool, string) {
	if p.cache_type == 0 {
		val, err := p.rdp.Get(p.ctx, key).Result()
		if err != nil {
			return false, ""
		}
		return true, val
	} else if p.cache_type == 1 {
		item, err := p.mc.Get(key)
		if err != nil {
			return false, ""
		}
		val := string(item.Value)
		return true, val

	}
	return false, ""
}
func (p *GCache) GetAndDelete(key string) (bool, string) {
	if p.cache_type == 0 {
		val, err := p.rdp.Get(p.ctx, key).Result()
		if err != nil {
			return false, ""
		}
		p.Delete(key)
		return true, val
	} else if p.cache_type == 1 {
		item, err := p.mc.Get(key)
		if err != nil {
			return false, ""
		}
		val := string(item.Value)

		p.Delete(key)
		return true, val
	}
	return false, ""

}
func (p *GCache) Delete(key string) bool {
	if p.cache_type == 0 {
		return p.Delete(key)
	} else if p.cache_type == 1 {
		if err := p.mc.Delete(key); err != nil {
			return false
		}
		return true
	}
	return false

}
