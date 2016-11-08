package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
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
	Ticktime      string  //数据时间 //原始数据结束
	LB5           float64 //前5天量比 平价每分钟的成交量
	LB            float64 //量比 当前量/分钟 /LB5
}

//是否已经计算5日量比
var haslb5 int

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

//数据排序
func order_by(args ...string) {
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
				return q.LB < p.LB // 量比 递减排序
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

//抓取最新数据
func catch_sina_list(ch chan string, page string) {
	url := `http://money.finance.sina.com.cn/d/api/openapi_proxy.php/?__s=[["hq","hs_a","volume",0,` + page + `,500]]`
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
	ch <- "ok:" + page

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	args := os.Args
	if len(args) == 2 {
		switch args[1] {
		case "1":
			list_init()
		case "2":
			order_by("")
		case "0":
			os.Exit(0)
		default:
		}
	}
	if len(args) == 1 {
		for {
			fmt.Println("操作目录: ")
			fmt.Println("1、抓取最新数据。")
			fmt.Println("2、成交量前十。2 1/2 10 (涨幅/成交量)")
			fmt.Println("3、test_run 抓取最新数据。")
			fmt.Println("8、量比选股大法。")
			fmt.Println("9、保存今日数据。")
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
				list_init()
			case "2":
				order_by(agrs...)
			case "8":
				LB_list()
			case "9":
				save_list()
			case "0":
				os.Exit(0)
			default:
				fmt.Println("default")
			}
			fmt.Println("-------处理完成-------")
		}
	}
}

//历史数据初始化
func list_init() {
	ch := make(chan string, 6)
	i := 1
	for i < 7 {
		go catch_sina_list(ch, strconv.Itoa(i))
		i++
	}
	for i > 1 {
		res := <-ch
		fmt.Println(res)
		i--
	}
	fmt.Println("数据采集完成。")
}

//量比选股
func LB_list() {
	e_time := time.Now().Unix() //当前时间戳
	db, _ := Mydb()
	if len(datalist) == 0 {
		list_init()
	}

	//计算5日量比数据
	if haslb5 == 0 {
		res, _ := _query(db, "SELECT `code`, `date` FROM `gu_day_history` WHERE `code`='000001' ORDER BY `date` DESC LIMIT 0,6")
		day := res[len(res)-1]["date"]
		//计算 5日成交量数据 每日4小时
		//分组统计前5天成交量之和
		tmp, _ := _query(db, "SELECT `code`, SUM(volume) as `v_all`  FROM `gu_day_history` WHERE `date`>'"+day+"' GROUP BY `code`")

		list_v_all := make(map[string]string)
		for _, tmp1 := range tmp {
			list_v_all[tmp1["code"]] = tmp1["v_all"]
		}
		for k, row := range datalist {
			datalist[k].LB5 = F64(list_v_all[row.Code]) / 240
		}
		haslb5 = 1
	}
	//计算量比
	the_time, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02")+" 09:25:00")
	s_time := the_time.Unix()     //开市时间戳
	fen := (e_time - s_time) / 60 //已开市分钟

	for key, row := range datalist {
		lb := fmt.Sprintf("%.3f", row.Volume/float64(fen)/row.LB5)
		datalist[key].LB = F64(lb)
	}
	order_by("2") //量比倒排输出

}

//保存数据
func save_list() {
	db, _ := Mydb()
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, -1)
	day := yesTime.Format("2006-01-02")

	//day := time.Now().Format("2006-01-02")
	//先删除
	_, err := db.Exec("DELETE FROM `gu_day_history` WHERE `date` = '" + day + "'")
	if err != nil {
		fmt.Println(err)
	}
	sql_str := "INSERT INTO `gu_day_history` (`code`, `date`, `open`, `close`, `volume`) VALUES "
	for _, row := range datalist {
		sql_str = sql_str + " ('" + row.Code + "','" + day + "'," + fmt.Sprintf("%.2f", row.Open) + "," + fmt.Sprintf("%.2f", row.Trade) + "," + fmt.Sprintf("%.2f", row.Volume) + "),"
	}
	end := len(sql_str) - 1
	in_sql := sql_str[0:end]

	//fmt.Println(in_sql)
	_, err = db.Exec(in_sql)
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
