// cleaner包,用于程序退出时,执行清理操作
package cleaner

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/syncmap"
)

// 定义默认清理列表
var (
	defaultList = &CleanList{list: new(syncmap.Map)}
)

// 定义清理器接口
type ExitCleaner interface {
	ExitClean() error // 清理方法签名，清理器实际执行的逻辑在该方法中实现
}

// 定义清理器
type Cleaner struct {
	F func() // 回调函数
}

// 实现清理器接口
// 在此方法中调用回调函数
func (p *Cleaner) ExitClean() {
	p.F()
}

// 添加回调函数
func (p *Cleaner) HandleFunc(f func()) {
	p.F = f
}

// 定义清理列表
type CleanList struct {
	list *syncmap.Map
}

// 将清理加入列表
func (p *CleanList) Add(key string, ec ExitCleaner) error {
	_, exist := p.list.Load(key)
	if exist {
		return errors.New("the key already exists")
	}
	p.list.Store(key, ec)
	return nil
}

// 清理列表的退出方法，调用传入的回调函数
// 清理器的签名方法将在回调函数中执行
func (p *CleanList) Exit(n int, f func(key interface{}, value interface{}) bool) {
	//创建一个元素为信号量的通道
	sigs := make(chan os.Signal, 1)

	//监听系统信号量
	signal.Notify(sigs,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
	go func(num int) {
		<-sigs
		p.list.Range(f)
		os.Exit(num)
	}(n)

}

// map的range方法的回调函数
// 用于运行ExitCleaner的Clean方法
func runClean(key interface{}, value interface{}) bool {
	newvalue, ok := value.(ExitCleaner)
	if !ok {
		return true
	}
	newvalue.ExitClean()
	return true
}

// 创建一个新的清理列表
func NewCleanList() *CleanList {
	return &CleanList{list: new(syncmap.Map)}
}

// 添加清理器到默认清理列表
func AddCleaner(key string, ec ExitCleaner) error {
	return defaultList.Add(key, ec)
}

// 默认清理列表退出函数
func Exit(n int) {
	defaultList.Exit(n, runClean)
}
