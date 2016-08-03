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
	rule map[string]string
	grep IGrep
}

func NewSpider(rule map[string]interface{}) *Spider {
	obj := new(Spider)
	tmp := make(map[string]string)
	tmp["list_url"] = rule["list_url"].(string)
	tmp["max_page"] = rule["max_page"].(string)
	tmp["cate"] = rule["cate"].(string)
	obj.rule = tmp

	switch t := rule["regexp"].(type) {
	default:
		log.Printf("unexpected type %T", t) // %T prints whatever type t has
	case *Grep1:
		obj.grep = rule["regexp"].(*Grep1)
	case *Grep2:
		obj.grep = rule["regexp"].(*Grep2)
	}
	return obj
}

//一个列表页处理
func (obj *Spider) do_list(url string) {
	index := obj.grep.Detail_url(url)
	//log.Println(index)
	db := Mydb()
	defer db.Close()
	for _, page := range index {
		body := obj.grep.Detail_content(page.url)
		InCar(db, page.title, body, obj.rule["cate"])
	}
}

/**
 * 多公司并发处理 (PT)
 * @param  {[type]} db *sql.DB       [description]
 * @return {[type]}    [description]
 */
func (obj *Spider) Run(c chan string) {
	max_page, _ := strconv.Atoi(obj.rule["max_page"])
	page := 0
	for i := 0; i < max_page; i++ {
		page = i + 1
		url := obj.grep.Page_url(obj.rule["list_url"], strconv.Itoa(page))
		obj.do_list(url)
		//log.Println(url)
	}
	c <- obj.rule["list_url"] + " done:" + strconv.Itoa(page)
}
