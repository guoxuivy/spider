package spider

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"runtime"
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
	results chan *goquery.Document
}

func NewPool() *Pool {
	obj := new(Pool)
	obj.jobs = make(chan string, 1000)
	obj.results = make(chan *goquery.Document, 1000)
	obj.Run()
	return obj
}

//添加任务
// func (obj *Pool) Add() {
// 	// 添加9个任务后关闭Channel
// 	// channel to indicate that's all the work we have.
// 	for j := 1; j <= 9; j++ {
// 		obj.jobs <- j
// 	}
// 	//close(obj.jobs)

// 	// //获取所有的处理结果
// 	for a := 1; a <= 9; a++ {
// 		<-obj.results
// 	}
// }

//添加任务
func (obj *Pool) Add(url string) {
	// 添加9个任务后关闭Channel
	obj.jobs <- url
}

func (obj *Pool) Res() *goquery.Document {
	dom := <-obj.results
	return dom
}

//这个是工作线程，处理具体的业务逻辑，将jobs中的任务取出，处理后将处理结果放置在results中。
func (obj *Pool) worker(id int) {
	for j := range obj.jobs {
		//log.Println("worker", id, "processing job", j)
		res, err := goquery.NewDocument(j)
		if err != nil {
			log.Println(err)
		}
		//time.Sleep(time.Second)
		obj.results <- res
	}
}

func (obj *Pool) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//两个channel，一个用来放置工作项，一个用来存放处理结果。
	// jobs := make(chan int, 100)
	// results := make(chan int, 100)

	// 开启三个线程，也就是说线程池中只有3个线程，实际情况下，我们可以根据需要动态增加或减少线程。
	for w := 1; w <= 20; w++ {
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

var Po *Pool

func init() {
	Po = NewPool()
}
