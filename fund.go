package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
)

var urlformat = "http://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=gp&rs=&gs=0&sc=rzdf&st=desc&sd=2019-12-10&ed=2020-12-10&qdii=&tabSubtype=,,,,,&pi=%d&pn=50&dx=1&v=0.08661388678343229"

//"http://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=zs&rs=&gs=0&sc=1yzf&st=desc&sd=2019-12-10&ed=2020-12-10&qdii=|&tabSubtype=,,,,,&pi=%d&pn=50&dx=1&v=0.9351668503260129"

//http://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=zs&rs=&gs=0&sc=1yzf&st=desc&sd=2019-12-10&ed=2020-12-10&qdii=|&tabSubtype=,,,,,&pi=3&pn=50&dx=1&v=0.9351668503260129
//"http://fund.eastmoney.com/data/rankhandler.aspx?op=ph&dt=kf&ft=gp&rs=&gs=0&sc=1yzf&st=desc&sd=2019-12-10&ed=2020-12-10&qdii=&tabSubtype=,,,,,&pi=%d&pn=50&dx=1&v=0.8597086393786593"
var referer = "http://fund.eastmoney.com/data/fundranking.html"
var max = 20.0

type Fund struct {
	Id    string
	Name  string
	Money float64
	Gains float64 //日涨跌
}

type FundSlice []Fund

var (
	Funds []Fund
	mutex sync.Mutex
	wg    sync.WaitGroup
)

func parse(id string) float64 {
	var url = fmt.Sprintf("http://fundf10.eastmoney.com/jdzf_%s.html", id)
	body := HttpReq(url)

	str := `\s*(.*?)亿元`
	r := regexp.MustCompile(str)
	matchs := r.FindStringSubmatch(string(body))

	if len(matchs) < 2 {
		return 0.0
	}
	v, _ := strconv.ParseFloat(string(matchs[1]), 64)
	return v
}

func main() {
	Funds = make([]Fund, 0)

	num := 3
	for i := 1; i < num; i++ {
		url := fmt.Sprintf(urlformat, i)
		body := HttpReq(url)

		reg := "[0-9]{6}"
		compile := regexp.MustCompile(reg)
		submatch := compile.FindAllSubmatch(body, -1)
		num := len(submatch)
		fmt.Println(num)

		for _, v := range submatch {
			//fmt.Printf("%s %.2f亿元\n",v[0], parse(string(v[0])))

			wg.Add(1)
			go Run(string(v[0]))
		}
	}

	wg.Wait()
	log.Println("end wait")

	sort.Sort(FundSlice(Funds))

	file, _ := os.OpenFile("dump.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	defer file.Close()
	for _, v := range Funds {
		//fmt.Printf("%v\n", v)
		b := fmt.Sprintf("%-7s %-20s %-7.2f %-7.2f\n", v.Id, v.Name, v.Money, v.Gains)
		file.Write([]byte(b))
	}
}

func Run(id string) {
	defer wg.Done()

	var url = fmt.Sprintf("http://fundf10.eastmoney.com/jdzf_%s.html", id)
	//log.Printf("go %s\n", url)

	body := HttpReq(url)

	str := `\s*(.*?)亿元`
	r := regexp.MustCompile(str)
	matchs := r.FindStringSubmatch(string(body))
	if len(matchs) < 2 {
		return
	}
	v, _ := strconv.ParseFloat(string(matchs[1]), 64)
	if v < max {
		return
	}

	str = `<title>(.*?)\(`
	r = regexp.MustCompile(str)
	matchs = r.FindStringSubmatch(string(body))
	title := matchs[1]

	str = `(.*?)([-]{0,}\d.*)\%`
	r = regexp.MustCompile(str)
	matchs = r.FindStringSubmatch(string(body))
	gains := matchs[2]
	f, _ := strconv.ParseFloat(gains, 32)

	mutex.Lock()
	Funds = append(Funds, Fund{id, title, v, f})
	mutex.Unlock()
}

func (a FundSlice) Len() int {
	return len(a)
}
func (a FundSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a FundSlice) Less(i, j int) bool {
	return a[j].Money < a[i].Money
}

/*HttpReq 发起http请求,返回body*/
func HttpReq(url string) []byte {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	//req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3314.0 Safari/537.36 SE 2.X MetaSr 1.0")
	req.Header.Add("Connection", "Close")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("http get error ", err)
		for {
			log.Println("retry", url)
			resp, err = client.Do(req)
			if err == nil {
				break
			}
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http ReadAll ", err)
		for {
			log.Println("retry ReadBody", url)
			resp, err = client.Do(req)
			if err == nil {
				body, err = ioutil.ReadAll(resp.Body)
				if err == nil {
					break
				}
			}
		}
	}

	return body
}

/*Save 保存文件*/
func Save(body []byte, filename string) error {
	err := ioutil.WriteFile(filename, body, 0666)
	if err != nil {
		fmt.Println("ioutil.WriteFile error", err)
		return err
	}
	fmt.Println(filename, " ok")
	return nil
}
