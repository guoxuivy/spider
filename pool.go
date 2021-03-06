package spider

import (
	"log"
	//"runtime"
	//"time"
)

/**
 *连接池实验
 *多路复用多线程
 *
 */

type Pool struct {
	//两个channel，一个用来放置工作项，一个用来存放处理结果。
	jobs    chan string
	results chan string
}

func NewPool() *Pool {
	obj := new(Pool)
	obj.jobs = make(chan string, 100)
	obj.results = make(chan string, 4000) //输出缓冲过小直接导致锁死
	obj.Run()
	log.Println("pool start")
	return obj
}

//添加任务
func (obj *Pool) Add(url string) {
	// 添加9个任务后关闭Channel
	obj.jobs <- url
}

func (obj *Pool) Res() string {
	dom := <-obj.results
	return dom
}

//这个是工作线程，处理具体的业务逻辑，将jobs中的任务取出，处理后将处理结果放置在results中。
func (obj *Pool) worker(id int) {
	for j := range obj.jobs {
		res := WaterImg(j)
		//log.Println("work", id)
		obj.results <- res
	}
}

func (obj *Pool) Run() {

	//两个channel，一个用来放置工作项，一个用来存放处理结果。
	// jobs := make(chan int, 100)
	// results := make(chan int, 100)

	// 开启三个线程，也就是说线程池中只有3个线程，实际情况下，我们可以根据需要动态增加或减少线程。
	for w := 1; w <= 100; w++ {
		go obj.worker(w)
	}

	// 添加9个任务后关闭Channel
	// channel to indicate that's all the work we have.
	// for j := 1; j <= 9; j++ {
	//  jobs <- j
	// }
	// close(jobs)

	// //获取所有的处理结果
	// for a := 1; a <= 9; a++ {
	//  <-results
	// }
}

//图片下载池
var Po *Pool

func init() {
	//Po = NewPool()
}
