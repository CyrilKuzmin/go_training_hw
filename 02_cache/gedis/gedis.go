package gedis

import (
	"sync"
	"time"
)

// Element struct contains value (interface{}) and its TTL
type element struct {
	value       interface{}
	validBefore time.Time
	ttl         time.Duration
}

type Gedis struct {
	data            map[string]*element
	cleanupInterval time.Duration
	sync.RWMutex
}

func (g *Gedis) Set(key string, value interface{}, ttl time.Duration) {
	g.Lock()
	defer g.Unlock()
	e := new(element)
	e.value = value
	e.ttl = ttl
	e.validBefore = time.Now().Add(ttl)
	g.data[key] = e
}

func (g *Gedis) Get(key string) (interface{}, bool) {
	g.Lock()
	defer g.Unlock()
	e, found := g.data[key]
	if !found {
		return nil, false
	}
	e.validBefore = time.Now().Add(e.ttl)
	return e.value, true
}

func (g *Gedis) Delete(key string) {
	g.Lock()
	defer g.Unlock()
	delete(g.data, key)
}

func (g *Gedis) getExpiredElements() (res []string) {
	g.Lock()
	defer g.Unlock()
	for k, v := range g.data {
		if time.Now().After(v.validBefore) {
			res = append(res, k)
		}
	}
	return
}

func cleanup(g *Gedis) {
	for {
		<-time.After(g.cleanupInterval)
		expired := g.getExpiredElements()
		if len(expired) != 0 {
			g.Lock()
			for _, k := range expired {

				delete(g.data, k)
			}
			g.Unlock()
		}

	}
}

func NewGedis(cleanupInterval time.Duration) *Gedis {
	data := make(map[string]*element)
	g := new(Gedis)
	g.cleanupInterval = cleanupInterval
	g.data = data
	go cleanup(g)
	return g
}
