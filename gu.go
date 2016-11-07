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
	"encoding/json"
	"errors"
	"sort"

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
参数说明：每页最多500条数据 6次获取即可 最多2940只A股 "volume",0,2,500] 按volume降序排列  规则为： 排序字段，排序，页数，每页数量
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

type SinaData struct {
	Uname         string
	Code          string
	Name          string
	Trade         float64 //最新价
	Pricechange   float64 //涨跌额
	Changepercent float64 //涨跌幅
	Buy           float64 //买入
	Sell          float64 //卖出
	Settlement    float64 //昨收
	Open          float64 //今开
	High          float64 //最高
	Low           float64 //最低
	Volume        float64 //成交量（手）
	Amount        float64 //成交额（万）
	Ticktime      string  //数据时间
}

//var datalist map[string]SinaData //无序map
var datalist []SinaData //有序排列

type ListWrap struct {
	datalist []SinaData
	by       func(p, q *SinaData) bool
}
type SortBy func(p, q *SinaData) bool

func (this ListWrap) Len() int { // 重写 Len() 方法
	return len(this.datalist)
}
func (this ListWrap) Swap(i, j int) { // 重写 Swap() 方法
	this.datalist[i], this.datalist[j] = this.datalist[j], this.datalist[i]
}
func (this ListWrap) Less(i, j int) bool { // 重写 Less() 方法
	return this.by(&this.datalist[i], &this.datalist[j])
}
func SortData(datalist []SinaData, by SortBy) { // SortPerson 方法
	sort.Sort(ListWrap{datalist, by})
}

func F64(a interface{}) float64 {
	res, err := strconv.ParseFloat(a.(string), 64)
	if err != nil {
		fmt.Println("float64数据错误：", a.(string))
		return 0
	}
	return res
}

//抓取最新数据 成交量倒排
func catch_sina_list(args ...string) {
	for p := 1; p <= 6; p++ {
		url := `http://money.finance.sina.com.cn/d/api/openapi_proxy.php/?__s=[["hq","hs_a","volume",0,` + strconv.Itoa(p) + `,500]]`
		resp, _ := http.Get(url)
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		//d := string(body)
		//去掉头尾[]
		jsonstr := body[1 : len(body)-1]
		var f interface{}
		err := json.Unmarshal(jsonstr, &f)
		if err != nil {
			fmt.Println("非json数据：", err)
		}

		m := f.(map[string]interface{})
		items := m["items"].([]interface{}) //数组断言
		for _, v := range items {
			row := v.([]interface{})
			code := row[1].(string)
			tmp := SinaData{}
			tmp.Uname = row[0].(string)
			tmp.Code = code
			tmp.Name = row[2].(string)
			tmp.Trade = F64(row[3])
			tmp.Pricechange = F64(row[4])
			tmp.Changepercent = F64(row[5])
			tmp.Buy = F64(row[6])
			tmp.Sell = F64(row[7])
			tmp.Settlement = F64(row[8])
			tmp.Open = F64(row[9])
			tmp.High = F64(row[10])
			tmp.Low = F64(row[11])
			tmp.Volume = F64(row[12])
			tmp.Amount = F64(row[13])
			tmp.Ticktime = row[14].(string)
			datalist = append(datalist, tmp)
		}
	}
}

//数据排序
func show_ord(args ...string) {
	args_num := len(args)
	if args_num == 0 {
		SortData(datalist, func(p, q *SinaData) bool {
			return q.Volume < p.Volume // 成交量 递减排序
			//return p.Changepercent < q.Changepercent // 递增排序
		})
	}
	if args_num > 0 {
		switch args[0] {
		case "1":
			SortData(datalist, func(p, q *SinaData) bool {
				return q.Changepercent < p.Changepercent // 涨幅 递减排序
			})
		case "2":
			SortData(datalist, func(p, q *SinaData) bool {
				return q.Volume < p.Volume // 成交额 递减排序
			})
		}
	}

	limit := 50
	if args_num == 2 {
		limit, _ = strconv.Atoi(args[1])
	}
	for i := 0; i < limit; i++ {
		fmt.Println(datalist[i])
	}

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
	args := os.Args
	if len(args) == 2 {
		switch args[1] {
		case "1":
			catch_sina_list("")
		case "2":
			show_ord("")
		case "0":
			os.Exit(0)
		default:
		}
	}
	if len(args) == 1 {
		for {
			fmt.Println("操作目录: ")
			fmt.Println("1、抓取最新数据。")
			fmt.Println("2、成交量前十。")
			fmt.Println("3、test_run 抓取最新数据。")
			fmt.Println("0、退出。 ")
			inputReader := bufio.NewReader(os.Stdin)
			command, _, _ := inputReader.ReadLine()
			code := string(command)
			code_arr := strings.Split(code, " ")
			//fmt.Println(code_arr)
			method := code_arr[0]
			agrs := code_arr[1:]
			switch method {
			case "1":
				catch_sina_list(agrs...)
			case "2":
				show_ord(agrs...)
			case "0":
				os.Exit(0)
			default:
				fmt.Println("default")
			}
			fmt.Println("-------处理完成-------")
		}
	}
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
	db, err := Mydb()
	if err != nil {
		fmt.Println(err)
	}
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
