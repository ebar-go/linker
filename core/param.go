package core

import "sync"

type Param struct {
	// This mutex protect Keys map
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}
}

func (p *Param) ResetKeys() {
	p.mu.Lock()
	p.Keys = make(map[string]interface{})
	p.mu.Unlock()
}

func (p *Param) Set(key string, value interface{}) {
	p.mu.Lock()
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}

	p.Keys[key] = value
	p.mu.Unlock()
}

func (p *Param) Get(key string) (value interface{}, exists bool) {
	p.mu.RLock()
	value, exists = p.Keys[key]
	p.mu.RUnlock()
	return

}

// GetInt returns the value associated with the key as an integer.
func (p *Param) GetInt(key string) (i int) {
	if val, ok := p.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetString returns the value associated with the key as a string.
func (c *Param) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}
