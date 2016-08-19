//car 爬虫
package spider

import (
	"log"
	"strconv"
)

type IGrep interface {
	//获取分页链接
	Page_url(url string, page string) string
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
}

func NewSpider(rule Guize) *Spider {
	obj := new(Spider)
	obj.rule = rule
	return obj
}

//一个列表页处理
func (obj *Spider) do_list(url string) {
	index := obj.rule.Grep.Detail_url(url)
	db, err := Mydb()
	if err != nil {
		log.Println(err)
	} else {
		//defer db.Close()
		for _, page := range index {
			body := obj.rule.Grep.Detail_content(page.url)
			InCar(db, page.title, body, obj.rule.Cate)
		}
	}
}

/**
 * 多公司并发处理 (PT)
 * @param  {[type]} db *sql.DB       [description]
 * @return {[type]}    [description]
 */
func (obj *Spider) Run(c chan string) {
	max_page, _ := strconv.Atoi(obj.rule.Max_page)
	page := 0
	for i := 0; i < max_page; i++ {
		page = i + 1
		url := obj.rule.Grep.Page_url(obj.rule.List_url, strconv.Itoa(page))
		obj.do_list(url)
		log.Println(url)
	}
	c <- obj.rule.List_url + " done:" + strconv.Itoa(page)
}
