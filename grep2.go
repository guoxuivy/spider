//car 爬虫
package spider

import (
	"log"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
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
	res.Find("#dasan_content").NextAll().Remove()
	res.Find("#dasan_content").Remove()
	content := res.Find(".content")
	//图片下载本地化
	//	content.Find("img").Each(func(i int, img *goquery.Selection) {
	//		src, _ := img.Attr("src")
	//		name, _ := GetImg(src)
	//		img.SetAttr("src", "/"+name)
	//	})
	body, _ = content.Html()
	return body
}
