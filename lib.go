package spider

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"crypto/md5"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
)

const (
	DSN string = "root:root@tcp(127.0.0.1:3306)/car?charset=utf8"
)

var _db *sql.DB

/**
 * 数据库连接
 */
func Mydb() (*sql.DB, error) {
	if _db != nil {
		return _db, nil
	}
	db, _ := sql.Open("mysql", DSN)
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(100)
	err := db.Ping()
	if err != nil {
		err = errors.New("数据库连接错误," + fmt.Sprint(DSN))
		return nil, err
	} else {
		_db = db
	}
	return _db, nil
}

/**
 * 升级日志写入  数据库保存
 * @param  {[type]} log string        [description]
 * @return {[type]}     [description]
 */
func InCar(db *sql.DB, title string, content string, cate string, url string) {
	urlmd5 := Md5(url)
	stmt, _ := db.Prepare("INSERT INTO `gai` (`title`, `content`, `cate`, `url`, `urlmd5`) VALUES (?,?,?,?,?)")
	defer stmt.Close()
	stmt.Exec(title, content, cate, url, urlmd5)
}

/**
 * url是否存在 检测
 * @param  {[type]} url string
 * @return bool
 */
func CheckUrl(db *sql.DB, url string) bool {
	urlmd5 := Md5(url)
	//查询数据库
	query, err := db.Query("SELECT count(*) as num FROM `gai` WHERE `urlmd5` = ? ", urlmd5)
	if err != nil {
		log.Println("查询数据库失败", err.Error())
	}
	defer query.Close()
	var num int
	for query.Next() {
		err = query.Scan(&num)
	}
	if num == 0 {
		return false
	} else {
		return true
	}

}

/**
 * 升级日志写入  文件追加参数奇葩 多 要3个
 * @param  {[type]} log string        [description]
 * @return {[type]}     [description]
 */
func WriteResult(tag string, data string) {
	str_time := time.Now().Format("2006_01_02")
	filename := tag + "_" + str_time + ".log"
	fl, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
	}
	defer fl.Close()
	fl.WriteString(data)
	fl.WriteString("\n")
}

func CleanBody(src string) string {
	//src := "<div style='eee'><script>style</script>ad<br/>fa<img src='fadfa'>df<p>wosp</p></div>"
	//将HTML标签全转换成小写
	re := regexp.MustCompile(`<[\S\s]+?\>`)
	src = re.ReplaceAllStringFunc(src, strings.ToLower)
	//去除STYLE
	re = regexp.MustCompile(`<style[\S\s]+?</style>`)
	src = re.ReplaceAllString(src, "")
	//去除SCRIPT
	re = regexp.MustCompile(`<script[\S\s]+?</script>`)
	src = re.ReplaceAllString(src, "")

	re = regexp.MustCompile(`</?[^/?(img)|(br)|(p)][^><]*>`)
	src = re.ReplaceAllString(src, "")
	//fmt.Printf("%q\n", re.FindAllString(src, -1))
	//fmt.Println(src)
	return src
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

//下载远程图片
func GetImg(url string) (file string, err error) {
	path := strings.Split(url, "/")
	var name string
	if len(path) > 1 {
		name = path[len(path)-1]
	}
	file = "update/" + time.Now().Format("20060102") + "/" + name
	dir, _ := os.Getwd()
	dir = dir + "/update/" + time.Now().Format("20060102")

	if isDirExists(dir) {
		//fmt.Println("目录存在")
	} else {
		err = os.MkdirAll(dir, os.ModePerm) //生成多级目录
		if err != nil {
			fmt.Println(err)
		}
	}

	//fmt.Println(name)
	out, err := os.Create(file)
	defer out.Close()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	_, err = io.Copy(out, bytes.NewReader(pix))
	return
}

func WaterFilelist(path string) {
	Po := NewPool()

	var list []string

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		//WaterImg(path)
		list = append(list, path)
		// Po.Add(path)
		// Po.Res()
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}

	// 添加9个任务后关闭Channel
	for j := 0; j < len(list); j++ {
		Po.Add(list[j])
	}

	//获取所有的处理结果
	for a := 0; a < len(list); a++ {
		Po.Res()
	}
}

func WaterImg(path string) string {
	//原始图片是sam.jpg
	imgb, _ := os.Open(path)
	img, _ := jpeg.Decode(imgb)
	defer imgb.Close()
	if img == nil {
		return ""
	}

	wmb, _ := os.Open("water.png")
	watermark, _ := png.Decode(wmb)
	defer wmb.Close()

	//把水印写到右下角，并向0坐标各偏移10个像素
	//offset := image.Pt(img.Bounds().Dx()-watermark.Bounds().Dx()-10, img.Bounds().Dy()-watermark.Bounds().Dy()-10)
	offset := image.Pt(1, img.Bounds().Dy()-watermark.Bounds().Dy()-1)
	b := img.Bounds()
	m := image.NewNRGBA(b)

	draw.Draw(m, b, img, image.ZP, draw.Src)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

	//生成新图片new.jpg，并设置图片质量..
	imgw, _ := os.Create("ok/" + path)
	jpeg.Encode(imgw, m, &jpeg.Options{100})
	defer imgw.Close()

	return "ok"
	//fmt.Println("水印添加结束,请查看new.jpg图片...")
}

func Md5(str string) (md5str string) {
	data := []byte(str)
	has := md5.Sum(data)
	md5str = fmt.Sprintf("%x", has)
	return
}
