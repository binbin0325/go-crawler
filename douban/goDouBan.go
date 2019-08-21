package douban

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/huichen/sego"
)

/*
* colly爬取 goQuery解析Dom sego分词  写入文件 利用微词云生成词云图
 */
func DouBanStart() {
	wordCount := make(map[int][]string)
	// 载入sego词典
	var segmenter sego.Segmenter
	segmenter.LoadDictionary("c:/gitlab/sego/data/dictionary.txt")
	// 构造colly
	c := colly.NewCollector()
	// 超时设定
	c.SetRequestTimeout(100 * time.Second)
	// 指定Agent信息
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36"
	craw(20, c, segmenter, wordCount)
	fmt.Println(wordCount)
}

func craw(start int, c *colly.Collector, segmenter sego.Segmenter, wordCount map[int][]string) {
	key := 1
	urlstr := "https://movie.douban.com/subject/26794435/comments?start=" + strconv.Itoa(start) + "&limit=20&sort=new_score&status=P"
	u, err := url.Parse(urlstr)
	if err != nil {
		log.Fatal(err)
	}
	//在发起请求前被调用
	c.OnRequest(func(r *colly.Request) {
		// Request头部设定
		r.Headers.Set("Host", u.Host)
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Origin", u.Host)
		r.Headers.Set("Referer", urlstr)
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Accept-Language", "zh-CN, zh;q=0.9")
		r.Headers.Set("Cookie", `bid=0O4PFLO6UDk; douban-fav-remind=1; ll="118124"; _pk_ref.100001.4cf6=%5B%22%22%2C%22%22%2C1566287759%2C%22https%3A%2F%2Fwww.google.com%2F%22%5D; _pk_ses.100001.4cf6=*; ap_v=0,6.0; __utma=30149280.1638934403.1566287760.1566287760.1566287760.1; __utmc=30149280; __utmz=30149280.1566287760.1.1.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided); __utma=223695111.1412641646.1566287760.1566287760.1566287760.1; __utmb=223695111.0.10.1566287760; __utmc=223695111; __utmz=223695111.1566287760.1.1.utmcsr=google|utmccn=(organic)|utmcmd=organic|utmctr=(not%20provided); __yadk_uid=y0ebDBi9yNX3GRlCEKinrp9MasZViL7H; trc_cookie_storage=taboola%2520global%253Auser-id%3D03d0f386-3ed3-4ebd-a6c4-d092f3244b83-tuct43d83fc; push_noty_num=0; push_doumail_num=0; __utmt_t1=1; dbcl2="76066885:s+alIFVNFis"; ck=JsK4; RT=s=1566290246052&r=https%3A%2F%2Fmovie.douban.com%2Fsubject%2F27010768%2Fcomments%3Fstart%3D320%26amp%3Bamp%3Bamp%3Blimit%3D20%26amp%3Bamp%3Bamp%3Bsort%3Dnew_score%26amp%3Bamp%3Bamp%3Bstatus%3DP; __utmb=30149280.41.8.1566290247228; _pk_id.100001.4cf6=5850da7fd5301309.1566287759.1.1566290247.1566287759.`)
	})
	//在OnResponse之后被调用，如果收到的内容是HTML
	c.OnHTML("title", func(e *colly.HTMLElement) {
		fmt.Println("title:", e.Text)
	})
	//收到回复后被调用
	c.OnResponse(func(resp *colly.Response) {
		fmt.Println("response received", resp.StatusCode)
		if start == 500 {
			return
		}
		// goquery直接读取resp.Body的内容
		htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body))
		if err != nil {
			log.Fatal(err)
		}
		// 找到抓取项 <div class="hotnews" alog-group="focustop-hotnews"> 下所有的a解析
		htmlDoc.Find(".comment p span").Each(func(i int, s *goquery.Selection) {
			title := s.Text()
			//分词
			participle(title, segmenter, wordCount, key)
			key++
		})
		start = start + 20
		time.Sleep(10 * time.Second)
		urlstr = "https://movie.douban.com/subject/26794435/comments?start=" + strconv.Itoa(start) + "&limit=20&sort=new_score&status=P"
		if err := c.Visit(urlstr); err != nil {
			fmt.Println(err.Error())
		}
	})
	//请求过程中如果发生错误被调用
	c.OnError(func(resp *colly.Response, errHttp error) {
		err = errHttp
	})
	//没有爬取任务时 执行
	c.OnScraped(func(r *colly.Response) {
		write(wordCount)
	})

	err = c.Visit(urlstr)
}

func participle(htmlContext string, segmenter sego.Segmenter, wordCount map[int][]string, key int) {
	// 分词
	text := []byte(htmlContext)
	segments := segmenter.Segment(text)
	// 处理分词结果
	// 支持普通模式和搜索模式两种分词，见代码中SegmentsToString函数的注释。
	//wordCountGO(sego.SegmentsToSlice(segments, false), wordCount)
	wordCount[key] = sego.SegmentsToSlice(segments, false)
	//fmt.Println(sego.SegmentsToString(segments, false))
}

func write(maps map[int][]string) {
	f, err := os.Create("c:/GoWork/go-crawler/wordCount.txt")
	if err != nil {
		fmt.Printf("create map file error: %v\n", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, v := range maps {
		lineStr := fmt.Sprintf("%s", v)
		fmt.Fprintln(w, lineStr)
	}
	w.Flush()
}
