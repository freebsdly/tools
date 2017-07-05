package system

import (
	"errors"

	"golang.org/x/sync/syncmap"
)

// 定义清理器接口
type ExitCleaner interface {
	ExitClean() error
}

// 定义清理器列表
type CleanList struct {
	list *syncmap.Map
}

//
func (p *CleanList) Add(ec ExitCleaner) error {
	_, exist := p.list.Load(ec)
	if exist {
		return errors.New("the key already exists")
	}
	p.list.Store(ec, "")
	return nil
}

//
func (p *CleanList) Run() {
	p.list.Range(runClean)
}

//
func runClean(key interface{}, value interface{}) bool {
	newkey, ok := key.(ExitCleaner)
	if !ok {
		return true
	}
	newkey.ExitClean()
	return true
}
