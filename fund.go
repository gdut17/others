package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"
)

//基金结构体
type Fund struct {
	Id    int
	Name  string
	Gains float64 //涨跌
}

type FundSlice []Fund

var (
	Funds FundSlice
	mutex sync.Mutex
	wg    sync.WaitGroup
)

func main() {
	file, err := os.Open("./mycode.txt")
	if err != nil {
		fmt.Println("read file failed")
		return
	}
	defer file.Close()

	Funds = make([]Fund, 0)
	var id int

	for {
		_, err := fmt.Fscanf(file, "%06d\n", &id)
		if err != nil {
			break
		}
		fmt.Printf("%06d\n", id)

		wg.Add(1)
		go Run(int(id))
	}
	wg.Wait()

	sort.Sort(FundSlice(Funds))

	//写入文件
	timeObj := time.Now()
	filename := fmt.Sprintf("./%d-%02d-%02d-%02d-%02d_merge.txt", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute())

	fmt.Println(filename)
	fp, err := os.Create(filename)
	if err != nil {
		fmt.Printf("error creating file: %v", err)
		return
	}
	defer fp.Close()

	for _, v := range Funds {
		msg := fmt.Sprintf("%06d %7.2f\t %-20s\n", v.Id, v.Gains, v.Name)
		fmt.Printf("%s", msg)
		_, err = fp.WriteString(msg)
		if err != nil {
			fmt.Printf("error writing string: %v", err)
		}
	}
}

func (a FundSlice) Len() int {
	return len(a)
}
func (a FundSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a FundSlice) Less(i, j int) bool {
	return a[j].Gains > a[i].Gains
}

func Run(id int) {
	url := fmt.Sprintf("http://fundf10.eastmoney.com/jdzf_%06d.html", id)
	defer wg.Done()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3314.0 Safari/537.36 SE 2.X MetaSr 1.0")
	req.Header.Add("Connection", "Close")
	req.Close = true

	response, err := client.Do(req)
	if err != nil {
		for {
			fmt.Println("retry", url)
			response, err = client.Do(req)
			if err == nil {
				break
			}
		}
	}

	if response.StatusCode != 200 {
		fmt.Println(url, response.StatusCode)
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		for {
			fmt.Println("retry", url)
			response, err = client.Do(req)
			if err == nil {
				body, err = ioutil.ReadAll(response.Body)
				if err == nil {
					break
				}
			}
		}
	}

	str := `<title>(.*?)\(`
	r := regexp.MustCompile(str)
	matchs := r.FindStringSubmatch(string(body))
	title := matchs[1]
	//fmt.Println(title)

	str = `(.*?)([-]{0,}\d.*)\%`
	r = regexp.MustCompile(str)
	matchs = r.FindStringSubmatch(string(body))
	gains := matchs[2]
	//fmt.Printf("%s\n", gains)

	f, _ := strconv.ParseFloat(gains, 32)

	//fmt.Printf("%06d %s %.2f \n", id, title,  f)

	mutex.Lock()
	Funds = append(Funds, Fund{id, title, f})
	mutex.Unlock()
}
