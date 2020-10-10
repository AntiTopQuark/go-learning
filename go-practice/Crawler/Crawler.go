package main

import (
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var base *url.URL

func main() {
	var maxDepth int
	var concurrency int
	var initalLink string

	flag.IntVar(&maxDepth, "d", 3, "最大爬取深度,默认为3")
	flag.IntVar(&concurrency, "c", 20, "最大并发访问量,默认为20")
	flag.StringVar(&initalLink, "u", "", "初始链接")

	flag.Parse()

	u, err := url.Parse(initalLink)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化链接错误:%s\n", err)
		os.Exit(1)
	}

	base = u
	parallellyCrawl(initalLink, concurrency, maxDepth)

}

type URLInfo struct {
	url   string
	depth int
}

func parallellyCrawl(initalLink string, concurrency int, depth int) {
	worklist := make(chan []URLInfo, 1)
	unessLinks := make(chan URLInfo, 1)
	// 值为1时表示进入unessLinks队列,值为2时表示爬虫完成了该页面
	seen := make(map[string]int)
	seenLock := sync.Mutex{}

	var urlInfos []URLInfo

	for _, url := range strings.Split(initalLink, " ") {
		urlInfos = append(urlInfos, URLInfo{url, 1})
	}

	// 1 将base 加入到wordlist
	go func() {
		worklist <- urlInfos
	}()

	// 2 死循环判断是否所有的url都执行完
	go func() {
		for {
			time.Sleep(2 * time.Second)
			seenFlag := true
			seenLock.Lock()

			for k := range seen {
				if seen[k] == 1 {
					seenFlag = false
				}
			}
			seenLock.Unlock()

			if seenFlag && len(worklist) == 0 {
				close(unessLinks)
				close(worklist)
				break
			}
		}
	}()

	// 3 开启多个爬虫协程,将爬取的链接放在wordlist里面
	for i := 0; i < concurrency; i++ {
		go func() {
			for link := range unessLinks {
				foundLinks := crawl(link.url)
				var urlInfos []URLInfo
				for _, u := range foundLinks {
					urlInfos = append(urlInfos, URLInfo{u, link.depth + 1})
				}
				go func(finishedUrl string) {
					worklist <- urlInfos
					seenLock.Lock()
					seen[finishedUrl] = 2
					seenLock.Unlock()
				}(link.url)
			}
		}()
	}

	for list := range worklist {
		for _, link := range list {
			if link.depth > depth {
				continue
			}
			seenLock.Lock()
			_, ok := seen[link.url]
			seenLock.Unlock()

			if !ok {
				seenLock.Lock()
				seen[link.url] = 1
				seenLock.Unlock()
				unessLinks <- link
			}
		}
	}
	fmt.Println("共访问了", len(seen), "个页面")

}

func crawl(u string) []string {
	list, err := Extract(u)

	if err != nil {
		log.Print(err)
	}
	return list
}

func Extract(url string) (urls []string, err error) {
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	response, err := client.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("getting %s:%s", url, response.StatusCode)
	}
	if err != nil {
		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
	}

	u, err := base.Parse(url)
	if err != nil {
		return nil, err
	}

	// TODO: 这里是干嘛的呢?
	if base.Host != u.Host {
		log.Printf("not saving %s: non-local", url)
		return nil, nil
	}
	var body io.Reader
	contentType := response.Header["Content-Type"]
	if strings.Contains(strings.Join(contentType, ","), "text/html") {
		doc, err := html.Parse(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("parsing %s as HTML: %v", u, err)
		}

		nodes := linkNodes(doc)

		urls = linkURLs(nodes, u)

		rewriteLocalLinks(nodes, u)

		b := &bytes.Buffer{}
		err = html.Render(b, doc)
		if err != nil {
			log.Printf("render %s: %s", u, err)
		}
		body = b
	}

	err = save(response, body)
	return urls, err

}

func rewriteLocalLinks(nodes []*html.Node, u *url.URL) {
	for _, n := range nodes {
		for i, a := range n.Attr {
			if a.Key != "href" {
				continue
			}
			link, err := base.Parse(a.Val)
			if err != nil || link.Host != base.Host {
				continue // ignore bad and non-local URLs
			}

			link.Scheme = ""
			link.Host = ""
			link.User = nil
			a.Val = link.String()

			n.Attr[i] = a
		}
	}
}

func linkNodes(n *html.Node) []*html.Node {
	var links []*html.Node
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			links = append(links, n)
		}
	}
	forEachNode(n, visitNode, nil)
	return links
}

func forEachNode(n *html.Node, pre, post func(n *html.Node)) {
	if pre != nil {
		pre(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, pre, post)
	}
	if post != nil {
		post(n)
	}
}
func linkURLs(linkNodes []*html.Node, base *url.URL) []string {
	var urls []string
	for _, n := range linkNodes {
		for _, a := range n.Attr {
			if a.Key != "href" {
				continue
			}
			link, err := base.Parse(a.Val)
			if err != nil {
				log.Printf("skipping %q: %s", a.Val, err)
				continue
			}
			if link.Host != base.Host {
				//log.Printf("skipping %q: non-local host", a.Val)
				continue
			}
			if strings.HasPrefix(link.String(), "javascript") {
				continue
			}
			urls = append(urls, link.String())
		}
	}
	return urls
}
func save(resp *http.Response, body io.Reader) error {
	u := resp.Request.URL

	filename := filepath.Join(u.Host, u.Path)

	if filepath.Ext(u.Path) == "" {
		filename = filepath.Join(u.Host, u.Path, "index.html")
	}

	err := os.MkdirAll(filepath.Dir(filename), 0777)

	if err != nil {
		return err
	}

	fmt.Println("filename:", filename)

	file, err := os.Create(filename)

	if err != nil {
		return err
	}

	if body != nil {
		_, err = io.Copy(file, body)
	} else {
		_, err = io.Copy(file, resp.Body)
	}

	if err != nil {
		log.Print("save: ", err)
	}

	err = file.Close()

	if err != nil {
		log.Print("save: ", err)
	}

	return nil
}
