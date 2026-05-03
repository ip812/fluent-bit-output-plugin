package main

import (
	"sync"
)

type Plugins struct {
	sync.Map // map[string]OutputPlugin
}

func NewPlugins() *Plugins {
	return &Plugins{}
}

func (p *Plugins) Contains(id string) bool {
	_, ok := p.Load(id)
	return ok
}

func (p *Plugins) Get(id string) (OutputPlugin, bool) {
	val, ok := p.Load(id)
	if !ok {
		return nil, false
	}

	plg, ok := val.(OutputPlugin)
	if !ok {
		return nil, false
	}

	return plg, ok
}

func (p *Plugins) Set(id string, op OutputPlugin) {
	p.Store(id, op)
}

func (p *Plugins) Remove(id string) {
	p.Delete(id)
}

func (p *Plugins) Len() int {
	count := 0
	p.Range(func(_, _ any) bool {
		count++
		return true
	})

	return count
}

func (p *Plugins) CleanupAll() {
	var idsToDelete []string

	p.Range(func(key, value any) bool {
		id, ok := key.(string)
		if !ok {
			return true
		}

		plg, ok := value.(OutputPlugin)
		if ok && p != nil {
			plg.Close()
		}

		idsToDelete = append(idsToDelete, id)

		return true
	})

	for _, id := range idsToDelete {
		p.Delete(id)
	}
}
