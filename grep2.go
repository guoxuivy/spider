//car 爬虫
package spider

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
)

type Grep2 struct {
}

//获取带分页的url
func (obj *Grep2) Page_url(url string, page string) string {
	re, _ := regexp.Compile(`_[\d]`)
	url = re.ReplaceAllString(url, "_"+page)
	return url
}

//获取详情页url
func (obj *Grep2) Detail_url(url string) []IndexItem {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
	}
	index := make([]IndexItem, 0)
	doc.Find("h4 a").Each(func(i int, li *goquery.Selection) {
		url, _ := li.Attr("href")
		title := li.Text()
		index = append(index, IndexItem{"http://www.wanchezhijia.com" + url, title})
	})
	return index
}

//详情处理
func (obj *Grep2) Detail_content(url string) string {
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
	//body = obj.clean(body)
	return body
}
