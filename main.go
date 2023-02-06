package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/idna"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"unicode"
)

type Yuming struct {
	Name     string
	State    string
	Height   int
	Highest  int
	Value    int
	Old_name string
}

type Person struct {
	Total  int
	Offset int
	Limit  int
	Result []Yuming
}

type Bids_name struct {
	Name string
	Bids []Bids
}
type Bids struct {
	Txid       string
	Lockup     int
	new_Lockup float64
}

var p *idna.Profile

func main() {
	file, err := os.OpenFile("b.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("open file failed, err:", err)
		return
	}
	limit := 20
	offset := 0
	fmt.Println(limit)
	fmt.Println(offset)
	fmt.Println(`----------`)
	fmt.Println(strconv.Itoa(limit))
	fmt.Println(strconv.Itoa(offset))
	//return 循环每一页
	for {
		apiUrl := "https://e.hnsfans.com/api/names?"
		//apiUrl := "https://e.hnsfans.com/api/names?limit=20&offset=0&status=bidding"
		// URL param
		data := url.Values{}
		data.Set("limit", strconv.Itoa(limit))
		data.Set("offset", strconv.Itoa(offset))
		//查看揭示
		//data.Set("status", "reveal")
		//查看出价
		data.Set("status", "bidding")
		u, err := url.ParseRequestURI(apiUrl)
		if err != nil {
			fmt.Printf("parse url requestUrl failed, err:%v\n", err)
		}
		u.RawQuery = data.Encode() // URL encode
		//fmt.Println(u.String())
		resp, err := http.Get(u.String())
		if err != nil {
			fmt.Printf("post failed, err:%v\n", err)
			return
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("get resp failed, err:%v\n", err)
			return
		}
		//fmt.Println(string(b))
		//str := `{
		//"total": 68652,
		//"offset": 0,
		//"limit": 50}`
		p1 := &Person{}
		//err := json.Unmarshal([]byte(str), p1)
		err = json.Unmarshal(b, p1)
		if err != nil {
			fmt.Println("json unmarshal failed!,", err)
			return
		}

		// 判断是否为中文。
		p = idna.New()
		//fmt.Println(p.ToUnicode("xn--i8s3qt32i"))
		// 首先循环里面的每一个，变为正常字符。
		for i := 0; i < len(p1.Result); i++ {
			p1.Result[i].Old_name = p1.Result[i].Name
			new_name, err := p.ToUnicode(p1.Result[i].Name)
			if err != nil {
				fmt.Println("转化中文域名错误请重试：,", err)
				continue
			}
			p1.Result[i].Name = new_name
		}
		//fmt.Printf("%#v\n", p1)

		// 判断是否是中文，是的话打印出来。
		for i := 0; i < len(p1.Result); i++ {
			if IsChineseChar(p1.Result[i].Name) {
				//获取报价
				new_strbids := get_baojia(p1.Result[i].Old_name)
				//fmt.Println(p1.Result[i].Name)
				file.WriteString(p1.Result[i].Name + new_strbids + "\n") //直接写入字符串数据
			}
		}
		if p1.Offset >= p1.Total {
			break
		}
		limit = p1.Limit
		offset = offset + limit

		defer func() {
			resp.Body.Close()
		}()
	}
	defer file.Close()
}

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
	}
	return false
}

// 获取域名当前报价
func get_baojia(name string) string {
	// 获取域名当前报价
	apiUrl := "https://e.hnsfans.com/api/names/" + name
	u, err := url.ParseRequestURI(apiUrl)
	if err != nil {
		fmt.Printf("parse url requestUrl failed, err:%v\n", err)
	}
	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("post failed, err:%v\n", err)
		return ""
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get resp failed, err:%v\n", err)
		return ""
	}

	bids_name1 := &Bids_name{}
	err = json.Unmarshal(b, bids_name1)
	if err != nil {
		fmt.Println("json unmarshal failed!,", err)
		return ""
	}
	//fmt.Println(bids_name1)
	new_strbids := ""
	for i := 0; i < len(bids_name1.Bids); i++ {
		bids_name1.Bids[i].new_Lockup = float64(bids_name1.Bids[i].Lockup / 1000000)
		new_strbids = new_strbids + "  ,  " + strconv.FormatFloat(bids_name1.Bids[i].new_Lockup, 'f', 2, 64)
	}
	//fmt.Println(new_strbids)
	return new_strbids
}
