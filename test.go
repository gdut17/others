package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	var id int = 161725
	s := GetFundInfo(id)
	fmt.Printf("%06d %s %.2f%% %.2f亿元 %s\n", s.Id,
		s.Name, s.Gains, s.Money, s.Union)

	id = 3834
	s = GetFundInfo(id)
	fmt.Printf("%06d %s %.2f%% %.2f亿元 %s\n", s.Id,
		s.Name, s.Gains, s.Money, s.Union)
}

type Fund struct {
	Id    int     // 基金编号
	Name  string  // 基金名称
	Gains float64 // 涨跌
	Money float64 // 资产
	Union string  // 所属基金
}

func httpreq(id int) []byte {
	url := fmt.Sprintf("http://fundf10.eastmoney.com/jdzf_%06d.html", id)
	//defer wg.Done()

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
		return nil
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
	return body
}

func GetFundInfo(Id int) Fund {

	body := httpreq(Id)
	if body == nil {
		return Fund{}
	}

	var reg string = `<title>(.*?)\(`
	compile := regexp.MustCompile(reg)
	matchs := compile.FindStringSubmatch(string(body))
	title := matchs[1]

	reg = `(.*?)([-]{0,}\d.*)\%`
	compile = regexp.MustCompile(reg)
	matchs = compile.FindStringSubmatch(string(body))
	gain, _ := strconv.ParseFloat(matchs[2], 32)

	reg = `\d{1,}\.\d\d亿元`
	compile = regexp.MustCompile(reg)
	matchsall := compile.FindAllSubmatch(body, -1)
	s := string(matchsall[0][0])
	money, _ := strconv.ParseFloat(s[:len(s)-7], 32)

	reg = `<a href="http://fund.eastmoney.com/company/[0-9]{3,}.html">.*?基金`
	compile = regexp.MustCompile(reg)
	matchsall = compile.FindAllSubmatch(body, -1)
	s = string(matchsall[0][0])
	unioner := s[strings.Index(s, ">")+1:]

	return Fund{
		Id:    Id,
		Name:  title,
		Gains: gain,
		Money: money,
		Union: unioner,
	}
}
