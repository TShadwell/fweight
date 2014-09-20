package cache

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const debug = false

//Cache implements a thread-safe cache for polling resources.
type Cache struct {
	GetValue   func() interface{}
	Delay      time.Duration
	NextUpdate time.Time
	value      interface{}
	error      error
	stale      bool
	sync.RWMutex
}

//Returns a Response with the current value of the Cache as an interface{}.
func (c *Cache) Get() interface{} {
	return c.Value()
}

//Returns a Response with the current value of the Cache.
func (c *Cache) Value() Response {
	c.RLock()
	defer c.RUnlock()
	if c.NextUpdate.Before(time.Now()) {
		if debug {
			log.Println("updating cache...", c.NextUpdate)
		}
		c.RUnlock()
		c.Lock()

		v := c.GetValue()
		if e, ok := v.(error); ok {
			c.error = e
			c.stale = true
		} else {
			c.value = v
			c.stale = false
			c.error = nil
		}

		c.NextUpdate = time.Now().Add(c.Delay)
		if debug {
			log.Println("next update in", c.NextUpdate)
		}
		c.Unlock()
		c.RLock()
	}

	return Response{
		Value: c.value,
		Stale: c.stale,
		Error: c.error,
	}
}

func New(get func() interface{}, delay time.Duration) (c *Cache) {
	c = new(Cache)
	c.GetValue = get
	c.Delay = delay
	return
}

type Response struct {
	Value interface{}
	Stale bool
	Error error
}

//HTTPResponse is a convenience function that makes the http request rq periodically and
//processes the response with 'process'.
func HTTPResponse(rq *http.Request, process func(b []byte) interface{}, client *http.Client, t time.Duration) *Cache {
	return New(func() interface{} {
		r, err := client.Do(rq)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		bt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		return process(bt)
	}, t)
}
