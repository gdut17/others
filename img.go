package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

/*http请求参数设置*/
var host string = "formaxfr.xyz"
var date string = "20200405"
var str string = "jcP3VEGv"
var flag string = "1000kb"

var origin string = "https://search.pstatic.net"
var referer string = "https://search.pstatic.net"

/*匹配个数*/
var num int
var wg sync.WaitGroup

//https://search.pstatic.net/common?src=https://i.imgur.com/Lto3oDz.jpg
var reg string = "img src=\"https://search.pstatic.net/common?src=https://i.imgur.com/[a-zA-Z0-9]{1,}.jpg"

var url = "http://picxxxx.top/2021/01/26/3040.html"

func main() {

	logFile, err := os.OpenFile("./down.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile) //设置输出位置

	start := time.Now()

	m3u8Body := HttpReq(url)
	//f, _ := os.OpenFile("1.html", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	ioutil.WriteFile("1.html", m3u8Body, 0644)

	// 正则匹配tsURL
	RegexpUrl(m3u8Body, reg)

	//wg.Wait()

	fmt.Printf("time: %s\n", time.Since(start))
}

/*HttpReq 发起http请求,返回body*/
func HttpReq(url string) []byte {

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3314.0 Safari/537.36 SE 2.X MetaSr 1.0")
	req.Header.Add("Connection", "Close")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http get error ", err)
		for {
			fmt.Println("retry", url)
			resp, err = client.Do(req)
			if err == nil {
				break
			}
		}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("http ReadAll ", err)
		for {
			fmt.Println("retry ReadBody", url)
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

/*RegexpUrl 正则匹配,go协程处理*/
func RegexpUrl(body []byte, reg string) int {

	compile := regexp.MustCompile(reg)
	submatch := compile.FindAllSubmatch(body, -1)
	num = len(submatch)

	log.Printf("match %d\n", num)

	if num == 0 {
		fmt.Println("no match")
		return 0
	}

	for k, v := range submatch {
		//url := fmt.Sprintf("https://%s/%s/%s/%s/hls/%s", host, date, str, flag, string(v[0]))

		//args[len(args)-1] = string(v[0])
		//url := fmt.Sprintf(format, args...)
		url := string(v[0])
		fmt.Printf("go %d %s\n", k, url)

		//wg.Add(1)
		//go GetTs(url, k)

		if k%50 == 0 {
			//wg.Wait()
		}
	}
	return num
}

/*GetTs 获取ts文件并保存*/
func GetTs(url string, k int) {
	defer wg.Done()

	filename := fmt.Sprintf("./ts/%d.jpg", k)
	file, err := os.Open(filename)
	if err == nil {
		fmt.Println(k, " exist")
		file.Close()
		return
	}

	body := HttpReq(url)
	Save(body, filename)
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
