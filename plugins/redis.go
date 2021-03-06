package plugins

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/johnwilson/webapp"
)

type PluginRedis struct {
	p *redis.Pool
}

func (r *PluginRedis) Init() error {
	rmi := int(webapp.Application.Config.Get("redis.max_idle").(int64))
	rit := time.Duration(webapp.Application.Config.Get("redis.idle_timeout").(int64)) * time.Second
	p := &redis.Pool{
		MaxIdle:     rmi,
		IdleTimeout: rit,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", webapp.Application.Config.Get("redis.server").(string))
			if err != nil {
				return nil, err
			}
			pw := webapp.Application.Config.Get("redis.password").(string)
			if len(pw) > 0 {
				if _, err := c.Do("AUTH", webapp.Application.Config.Get("redis.password").(string)); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	// test connection
	_, err := p.Dial()
	if err != nil {
		return fmt.Errorf("redis: redis connection failed:\n%s", err)
	}

	r.p = p
	return nil
}

func (r *PluginRedis) Get() interface{} {
	return r.p
}

func (r *PluginRedis) Close() error {
	if err := r.p.Close(); err != nil {
		return fmt.Errorf("redis: redis connection close failed:\n%s", err)
	}
	return nil
}
