//car 爬虫
package spider

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
)

type Grep2 struct {
}

//获取带分页的url
func (obj *Grep2) Page_url(url string, page int) string {
	re := regexp.MustCompile(`_([\d]+)`)
	p := re.FindStringSubmatch(url)
	start_page, _ := strconv.Atoi(p[1])
	new_page := start_page + page
	url = re.ReplaceAllString(url, "_"+strconv.Itoa(new_page))
	return url
}

//获取详情页url
func (obj *Grep2) Detail_url(url string) []IndexItem {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
	}
	// Po.Add(url)
	// doc := Po.Res()

	index := make([]IndexItem, 0)
	doc.Find("dd h4 a").Each(func(i int, li *goquery.Selection) {
		url, _ := li.Attr("href")
		title := li.Text()
		index = append(index, IndexItem{"http://www.wanchezhijia.com" + url, title})
	})
	return index
}

//详情处理
func (obj *Grep2) Detail_content(url string) string {
	body := ""
	res, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
		return ""
	}
	// Po.Add(url)
	// res := Po.Res()

	content := res.Find("#dasan_content").PrevAll()
	content.Each(func(i int, p *goquery.Selection) {
		tmp, _ := p.Html()
		body = tmp + body
	})
	//log.Println(body)
	return body
}
