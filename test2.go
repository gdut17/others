package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	body, _ := ioutil.ReadFile("fund.html")
	str := `\d{1,}\.\d\d亿元`
	r := regexp.MustCompile(str)
	matchs := r.FindAllSubmatch(body, -1) //r.FindStringSubmatch(string(body))
	//money := matchs[0]
	fmt.Println(len(matchs))
	fmt.Printf("%s\n", matchs[0][0])
}
func main1() {

	str := "20.80亿元"
	fmt.Println(len(str))
	money, _ := strconv.ParseFloat(str[:len(str)-7], 32)
	fmt.Printf("%.2f\n", money)

	s := "<a href=\"http://fund.eastmoney.com/company/80000222.html\">华夏基金"
	fmt.Printf("%s\n", s[strings.Index(s, ">")+1:])
}
