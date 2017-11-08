//
package timer

//
type Timer interface {
	Add(j Jober, period uint32, onetime bool) (jid uint32, err error)
	Remove(jid uint32) error
	Start() error
}

// caller implement this interface
type Jober interface {
	Run() error
	Stop() error
}
