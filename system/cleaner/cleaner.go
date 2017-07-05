// cleaner包,用于程序退出时,执行清理操作
package cleaner

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/syncmap"
)

//
var (
	defaultList = &CleanList{list: new(syncmap.Map)}
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

// map的range方法的回调函数
// 用于运行ExitCleaner的Clean方法
func runClean(key interface{}, value interface{}) bool {
	newkey, ok := key.(ExitCleaner)
	if !ok {
		return true
	}
	newkey.ExitClean()
	return true
}

//
func AddCleaner(ec ExitCleaner) error {
	return defaultList.Add(ec)
}

//
func Exit(n int) {
	//创建一个元素为信号量的通道
	sigs := make(chan os.Signal, 1)

	//监听系统信号量
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	go func(num int) {
		<-sigs
		defaultList.Run()
		os.Exit(num)
	}(n)

}
