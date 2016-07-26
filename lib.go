package spider

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const (
	DSN string = "root:root@tcp(127.0.0.1:3306)/car?charset=utf8"
)

/**
 * 数据库连接
 */
func Mydb() *sql.DB {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Fatalf("Open database error: %s\n", err)
	}
	//defer db.Close() //不关闭连接
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

/**
 * 升级日志写入  数据库保存
 * @param  {[type]} log string        [description]
 * @return {[type]}     [description]
 */
func InCar(db *sql.DB, title string, content string, cate string) {
	stmt, _ := db.Prepare("INSERT INTO `gai` (`title`, `content`, `cate`) VALUES (?,?,?)")
	defer stmt.Close()
	stmt.Exec(title, content, cate)
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
