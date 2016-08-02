//car 爬虫
package spider

import (
	//"fmt"
	"log"
	//"net/url"
	"regexp"
	"strconv"
	//"time"
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

/**
 * 更新一个公司的客户状态 (PT) 考虑新建数据库连接 提高效率
 * @param  {[type]} db      *sql.DB       [description]
 * @param  {[type]} c       chan          int           [description]
 * @param  {[type]} comp_id int           [公司ID]
 * @param  {[type]} num int           	  [有效期天数]
 * @return {[type]}         			  [description]
 */
func (obj *Spider) clean(body string) (str string) {
	src := string(body)
	//将HTML标签全转换成小写
	// re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	// src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//去除STYLE
	re, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//去除SCRIPT
	re, _ = regexp.Compile("<script[\\S\\s]+?</script>")
	src = re.ReplaceAllString(src, "")

	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ = regexp.Compile("</?[^/?(img)|(br)][^><]*>")
	src = re.ReplaceAllString(src, "")

	//img 处理
	re, _ = regexp.Compile(`w=`)
	src = re.ReplaceAllString(src, "width=")

	re, _ = regexp.Compile(`src="data`)
	src = re.ReplaceAllString(src, `src="http://www.gai001.com/data`)

	//去除连续的换行符
	re, _ = regexp.Compile(`[\s]{2,}`)
	src = re.ReplaceAllString(src, "\n")

	return src
}

//一个列表页处理
func (obj *Spider) do_list(url string) {
	index := obj.grep.Detail_url(url)
	log.Println(index)
	// db := Mydb()
	// defer db.Close()
	// for _, page := range index {
	// 	body := obj.do_detail(page.url)
	// 	body := obj.grep.Detail_content(url)
	// 	InCar(db, page.title, body, obj.rule["cate"])
	// }
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
	}
	c <- obj.rule["list_url"] + " done:" + strconv.Itoa(page)
}
