package system


// 定义清理器接口
type ExitCleaner interface {
	ExitClean() error
}

// 定义清理器列表
type CleanList struct {
		
}

//
func (p *CleanList) Add(ec ExitCleaner) error {

}

//
func
