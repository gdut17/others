/*
	2021.01.04
*/
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

/*http请求参数设置*/
var host string
var date string
var str string
var flag string

var origin string
var referer string

/*匹配个数*/
var num int
var wg sync.WaitGroup

/*AES key*/
var aeskey []byte

/*需要修改的参数*/
var m3u8Url = "http://127.0.0.1/1.m3u8"
var reg = "[a-z0-9]{6}.ts"

const format = `https://%s/%s/%s/%s/hls/%s`

var ts string
var args []interface{} = []interface{}{host, date, str, flag, ts}

func main() {
	os.Mkdir("ts", os.ModePerm)
	os.Mkdir("merge", os.ModePerm)

	logFile, err := os.OpenFile("./down.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile) //设置输出位置

	m3u8FileName := "1.m3u8"
	m3u8Body := HttpReq(m3u8Url)
	Save(m3u8Body, m3u8FileName)

	num = RegexpUrl(m3u8Body, reg)
	log.Printf("match %d %s\n", m3u8Url, num)
	if num == 0 {
		return
	}

	wg.Wait()
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

	if num == 0 {
		fmt.Println("no match")
		return 0
	}

	for k, v := range submatch {
		//url := fmt.Sprintf("https://%s/%s/%s/%s/hls/%s", host, date, str, flag, string(v[0]))

		args[len(args)-1] = string(v[0])
		url := fmt.Sprintf(format, args...)
		fmt.Printf("go %d %s\n", k, url)

		wg.Add(1)
		go GetTs(url, k)

		if k%50 == 0 {
			wg.Wait()
		}
	}
	return num
}

/*GetTs 获取ts文件并保存*/
func GetTs(url string, k int) {
	defer wg.Done()

	filename := fmt.Sprintf("./ts/%d.ts", k)
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

/*MergeTs 合并ts文件*/
func MergeTs(num int, isAES bool) {
	tsFile := fmt.Sprintf("./merge/all_%s_%s.ts", date, str)
	file, err := os.Open(tsFile)
	if err == nil {
		fmt.Println(tsFile, "MergeTs exist")
		file.Close()
		return
	}

	fii, err := os.OpenFile(tsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fii.Close()

	for i := 0; i < num; i++ {

		fname := fmt.Sprintf("./ts/%d.ts", i)
		f, err := os.OpenFile(fname, os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
			return
		}

		if isAES == true {
			b, err = AesDecrypt(b, aeskey)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		//追加写入
		fii.Write(b)
		f.Close()

		//删除文件
		os.Remove(fname)

		fmt.Println(i, "ok")
	}
	fmt.Println("done")
}

/*AesDecrypt AES-128J解密*/
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	return origData, nil
}
