package main

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func videoDLWorker(destFile string, target string) error {
	resp, err := http.Get(target)
	if err != nil {
		log.Println(fmt.Sprintf("Http.Get\nerror: %s\ntarget: %s\n", err, target))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println(fmt.Sprintf("reading answer: non 200[code=%v] status code received: '%v'",
			resp.StatusCode, err))
		return errors.New("non 200 status code received")
	}
	err = os.MkdirAll(filepath.Dir(destFile), 0755)
	if err != nil {
		return err
	}
	if fileExists(destFile) {
		return nil
	}

	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println(fmt.Sprintln("download video err=", err))
		return err
	}
	return nil
}

func main() {
	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{DomainGlob: "*.down.icharle.*", Parallelism: 3})
	extensions.RandomUserAgent(c)

	detailLink := c.Clone()
	c.OnHTML(".mdui-row > ul > li", func(e *colly.HTMLElement) {

		name := e.DOM.Find("span").Text()

		fmt.Println("name : ", name)
		link := e.ChildAttr("a", "href")
		linkURL := fmt.Sprintf("%s://%s%s", e.Request.URL.Scheme, e.Request.URL.Hostname(), link)
		fmt.Println("link : ", linkURL)
		if link != "" {
			detailLink.Visit(linkURL)
		}
	})

	detailLink.OnHTML(".mdui-row > ul > li > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		name := e.ChildText("span")

		if strings.HasSuffix(link, ".mp4") {
			fmt.Println("name:", name)
			downLoadURL := fmt.Sprintf("%s://%s%s", e.Request.URL.Scheme, e.Request.URL.Hostname(), link)
			fmt.Println("link :", downLoadURL)
			err := videoDLWorker(name, downLoadURL)
			if err != nil {
				log.Println(err)
			}
		}
	})

	c.Visit("https://down.icharle.com/?/Go%E8%AF%AD%E8%A8%80%E5%AE%9E%E6%88%98%E6%B5%81%E5%AA%92%E4%BD%93%E8%A7%86%E9%A2%91%E7%BD%91%E7%AB%99/")

}
