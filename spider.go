//car 爬虫
package spider

import (
	//"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"regexp"
	"strconv"
	//"time"
)

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
}

func NewSpider(rule map[string]string) *Spider {
	obj := new(Spider)
	obj.rule = rule
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

//详情处理
func (obj *Spider) do_detail(url string) (str string) {
	res, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
	}
	content := res.Find("#article_content")
	content.Find("img").Each(func(i int, img *goquery.Selection) {
		img.RemoveAttr("onclick").RemoveAttr("zoomfile").RemoveAttr("id").RemoveAttr("aid").RemoveAttr("title").RemoveAttr("alt").RemoveAttr("onmouseover")
		src, _ := img.Attr("src")
		file, _ := img.Attr("file")
		if len(src) == 0 {
			img.SetAttr("src", file)
		}
		img.RemoveAttr("file")
	})
	content.Find("ins").Remove()
	content.Find("p").Last().Remove()

	body, _ := content.Html()
	body = obj.clean(body)
	return body
}

//一个列表页处理
func (obj *Spider) do_list(url string) {
	db := Mydb()
	defer db.Close()
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
	}
	index := make([]IndexItem, 0)
	doc.Find(".xs2 .xi2").Each(func(i int, li *goquery.Selection) {
		url, _ := li.Attr("href")
		title := li.Text()
		index = append(index, IndexItem{url, title})
	})
	//log.Println(index)
	for _, page := range index {
		body := obj.do_detail(page.url)
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
	uri, err := url.Parse(obj.rule["list_url"])
	if err != nil {
		log.Println(err)
	}
	q := uri.Query()
	page := 0
	for i := 0; i < max_page; i++ {
		page = i + 1
		q.Set(obj.rule["page_tag"], strconv.Itoa(page))
		uri.RawQuery = q.Encode()
		obj.do_list(uri.String())
	}
	c <- obj.rule["list_url"] + " done:" + strconv.Itoa(page)
}
