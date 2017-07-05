/*
	pool go程池
*/
package pool

import (
	"sync"
)

// 定义go程池结构体
type taskPool struct {
	Queue chan int
	Wg    *sync.WaitGroup
}

//
var (
	TaskPool taskPool
	TaskNum  int = 1000 //定义池最大容量
)

// 初始化go程池
func (this *taskPool) Init(maxnum, total int) {
	this.Queue = make(chan int, maxnum)
	this.Wg = new(sync.WaitGroup)
	this.Wg.Add(total)
}

// 添加任务到池
func (this *taskPool) AddTask() {
	this.Queue <- 1
}

// 从池中删除任务
func (this *taskPool) DeleteTask() {
	<-this.Queue
	this.Wg.Done()
}
