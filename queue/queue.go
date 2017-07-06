// queue包提供一个队列接口，后端可以用多种方式实现阻塞/非阻塞队列
package queue

type Queuer interface {
	Push()
	Pull()
}
