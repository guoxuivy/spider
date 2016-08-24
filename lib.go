package spider

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"bytes"
	"io"
	"io/ioutil"
	"net/http"
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
	stmt, _ := db.Prepare("INSERT INTO `gai` (`title`, `content`, `cate`, `url`) VALUES (?,?,?,?)")
	defer stmt.Close()
	stmt.Exec(title, content, cate, url)
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
		err = os.MkdirAll(dir+"/update/"+time.Now().Format("20060102"), os.ModePerm) //生成多级目录
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
