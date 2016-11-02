package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"

	"bufio"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

const (
	DSN string = "root:root@tcp(127.0.0.1:3306)/car?charset=utf8"
)

var _db *sql.DB

/**
接口整理

1、新浪批量实时数据接口
页面地址：http://finance.sina.com.cn/data/#stock
接口地址：http://money.finance.sina.com.cn/d/api/openapi_proxy.php/?__s=[["hq","hs_a","volume",1,2,500]]
参数说明：每页最多500条数据 6次获取即可 最多2940只A股 "volume",1,2,500] 按volume降序排列  规则为： 排序字段，排序，页数，每页数量
返回结构说明：
["symbol","code",		"name",				"trade","pricechange","changepercent","buy","sell","settlement","open","high","low","volume","amount",   			   "ticktime","per","per_d","nta","pb","mktcap","nmc","turnoverratio","favor","guba"]
			代码			名称						最新价	涨跌额	涨跌幅	买入		卖出		昨收		今开		最高		最低		成交量（手）	成交额（万）  数据时间
["sh600408","600408","\u5b89\u6cf0\u96c6\u56e2","5.930","0.310","5.516","5.930","5.940","5.620","5.770","6.180","5.710","157431523","943257675","11:30:00",148.25,-95.185,"1.0973",5.404,597032.4,597032.4,15.63682,"",""],

2、
实时数据接口
http://hq.sinajs.cn/list=sz002609
历史数据接口
http://data.gtimg.cn/flashdata/hushen/weekly/sz002609.js 周数据
http://money.finance.sina.com.cn/corp/go.php/vMS_MarketHistory/stockid/601006.phtml?year=2016&jidu=3

http://table.finance.yahoo.com/table.csv?s=000001.sz


*/
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
func InGu(db *sql.DB, code string, cate string) {
	//LOAD DATA LOCAL INFILE 'C:\\Users\\Administrator\\Desktop\\222222222\\sz002609.csv' INTO TABLE `car`.`gu_history` FIELDS ESCAPED BY '\\' TERMINATED BY ',' LINES TERMINATED BY '\n' (`date`, `open`, `high`, `low`, `close`, `volume`, `adj_close`);
	//UPDATE `car`.`gu_history` SET `code` = '000001' WHERE `code` = '';
	//DELETE FROM gu_history WHERE `date`='0000-00-00'

	filePath := "gu_csv/" + code + ".csv"
	mysql.RegisterLocalFile(filePath)
	sql := `LOAD DATA LOCAL INFILE '` + filePath + `' INTO TABLE gu_history FIELDS ESCAPED BY '\\' TERMINATED BY ',' LINES TERMINATED BY '\n' ` + " (`date`, `open`, `high`, `low`, `close`, `volume`, `adj_close`)"
	_, err := db.Exec(sql)
	if err != nil {
		fmt.Println(err)
	}
	mysql.DeregisterLocalFile(filePath)
	//	_, err = db.Exec("UPDATE `gu_history` SET `code` = '" + code + "' WHERE `code` = ''")
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	_, err = db.Exec("DELETE FROM gu_history WHERE `date`='0000-00-00'")
	//	if err != nil {
	//		fmt.Println(err)
	//	}
}

/**
 * 升级日志写入  文件追加参数奇葩 多 要3个
 * @param  {[type]} log string        [description]
 * @return {[type]}     [description]
 */
func WriteResult(code string, data string) {
	filename := "gu_csv/" + code + ".csv"
	os.Remove(filename) //删除文件test.txt
	fl, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer fl.Close()
	fl.WriteString(data)
	fl.WriteString("\n")
}

func do_one(db *sql.DB, code string, cate string) {
	resp, _ := http.Get("http://table.finance.yahoo.com/table.csv?s=" + code + "." + cate)
	fmt.Println("http://table.finance.yahoo.com/table.csv?s=" + code + "." + cate)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	r, _ := utf8.DecodeRune(body) // 解码 b 中的第一个字符
	//first := fmt.Print("%c\n", r) // 显示读出的字符
	if string(r) == "<" {
		fmt.Println("404")
		stmt, _ := db.Prepare("UPDATE `gu` SET `status`='0' WHERE `code` = ?")
		defer stmt.Close()
		stmt.Exec(code)
		return
	}
	d := string(body)
	WriteResult(code, d)
	get_sql(db, code)

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gu_init()
}

//分页参数为500
func do_limit(db *sql.DB, ch chan string, offset string) {
	res, _ := findAll(db, offset, "300")
	//return
	//fmt.Println(res)
	if res != nil {
		for _, v := range res {
			code := v["code"]
			cate := v["cate"]
			if cate == "sh" {
				cate = "ss"
			}
			do_one(db, code, cate)
		}
	}

	ch <- "ok" + offset + ":" + strconv.Itoa(len(res))
}

//历史数据初始化
func gu_init() {
	//下载cvs文件
	db, _ := Mydb()
	ch := make(chan string, 10)
	i := 0
	for i < 10 {
		offset := i * 300
		go do_limit(db, ch, strconv.Itoa(offset))
		i++
	}
	for i > 0 {
		res := <-ch
		fmt.Println(res)
		i--
	}
}

func ReadLine(sql *string, code string, fileName string, handler func(string, *string, string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line, sql, code)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}
func PrintSql(line string, str *string, code string) {
	if line != "" {
		arr := strings.Split(line, ",")
		for i, v := range arr {
			arr[i] = "'" + v + "'"
		}
		nline := strings.Join(arr, ",")
		*str = *str + " ('" + code + "'," + nline + "),"
	}
}

func get_sql(db *sql.DB, code string) {
	filePath := "gu_csv/" + code + ".csv"
	sql_str := "INSERT INTO `gu_history` (`code`, `date`, `open`, `high`, `low`, `close`, `volume`, `adj_close`) VALUES "
	ReadLine(&sql_str, code, filePath, PrintSql)
	end := len(sql_str) - 1
	in_sql := string(sql_str[0:end])
	//fmt.Println(in_sql)
	_, err := db.Exec(in_sql)
	if err != nil {
		fmt.Println(err)
	}
}

//查询全部任务
func findAll(db *sql.DB, offset string, size string) (map[int]map[string]string, error) {
	return _query(db, "SELECT * FROM `gu_1` WHERE `status`=1 ORDER BY id ASC LIMIT "+offset+","+size)
}

//通用列表查询
func _query(db *sql.DB, sql string) (map[int]map[string]string, error) {
	//查询数据库
	query, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer query.Close()

	//读出查询出的列字段名
	cols, _ := query.Columns()
	//values是每个列的值，这里获取到byte里
	values := make([][]byte, len(cols))
	//query.Scan的参数，因为每次查询出来的列是不定长的，用len(cols)定住当次查询的长度
	scans := make([]interface{}, len(cols))
	//让每一行数据都填充到[][]byte里面
	for i := range values {
		scans[i] = &values[i]
	}

	//最后得到的map
	results := make(map[int]map[string]string)
	i := 0
	for query.Next() {
		//query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
		if err := query.Scan(scans...); err != nil {
			return nil, err
		}
		row := make(map[string]string) //每行数据
		for k, v := range values {     //每行数据是放在values里面，现在把它挪到row里
			key := cols[k]
			row[key] = string(v)
		}
		results[i] = row //装入结果集中
		i++
	}
	return results, nil
}
