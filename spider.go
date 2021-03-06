//car 爬虫
package spider

import (
	"database/sql"
	"log"
	"strconv"
)

type IGrep interface {
	//获取分页链接
	Page_url(url string, page int) string
	//获取分页链接
	Detail_url(url string) []IndexItem

	Detail_content(url string) string
}

//采集器配置规则
type Guize struct {
	List_url string
	Max_page string
	Cate     string
	Grep     IGrep
}

//列表结构
type IndexItem struct {
	url   string
	title string
}

/**
 *公用蜘蛛
 * 一只蜘蛛负责爬行一个列表页面，带分页参数
 **/
type Spider struct {
	rule Guize
	db   *sql.DB
}

func NewSpider(rule Guize) *Spider {
	obj := new(Spider)
	obj.rule = rule
	return obj
}

//一个列表页处理
func (obj *Spider) do_list(url string) {
	index := obj.rule.Grep.Detail_url(url)
	c := make(chan int, 10)
	for _, page := range index {
		//高并发处理
		go func(page IndexItem) {
			if false == CheckUrl(obj.db, page.url) {
				body := obj.rule.Grep.Detail_content(page.url)
				InCar(obj.db, page.title, body, obj.rule.Cate, page.url)
				log.Println(page.url)
			}
			c <- 1
		}(page)
	}
	//获取所有的处理结果 确保无丢失
	for a := 1; a <= len(index); a++ {
		<-c
	}
}

/**
 * 多公司并发处理 (PT)
 * @param  {[type]} db *sql.DB       [description]
 * @return {[type]}    [description]
 */
func (obj *Spider) Run(c chan string) {
	db, err := Mydb()
	//defer db.Close()
	if err != nil {
		log.Println(err)
	}
	obj.db = db

	max_page, _ := strconv.Atoi(obj.rule.Max_page)
	page := 0
	for i := 0; i < max_page; i++ {
		page = i + 1
		url := obj.rule.Grep.Page_url(obj.rule.List_url, page)
		obj.do_list(url)
	}
	c <- obj.rule.List_url + " done:" + strconv.Itoa(page)
}
