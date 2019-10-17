package main

import (
	"errors"
	"flag"
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

const usageString string = `Usage:
	crawlerTea [flags] 
	
	Download a video from URL.
	Example: crawlerTea -i https://down.icharle.com/?/Go语言实战流媒体视频网站/ -d ./GolangMedia
Flags:`

func main() {
	flag.Usage = func() {
		fmt.Println(usageString)
		flag.PrintDefaults()
	}
	var outputDir string
	var tmpDir string
	var targetUrl string

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&outputDir, "d", currentDir, "The output directory.")
	flag.StringVar(&targetUrl, "i", "", "Target URL")
	flag.Parse()
	log.Println(flag.Args())

	fmt.Println("outputDir:", outputDir)

	c := colly.NewCollector()
	c.Limit(&colly.LimitRule{DomainGlob: "*.down.icharle.*", Parallelism: 3})
	extensions.RandomUserAgent(c)

	detailLink := c.Clone()
	c.OnHTML(".mdui-row > ul > li", func(e *colly.HTMLElement) {

		name := e.DOM.Find("span").Text()
		name = strings.Replace(name, " ", "", -1)
		if name != "" {
			fmt.Printf("%s\n", name)
			link := e.ChildAttr("a", "href")
			linkURL := fmt.Sprintf("%s://%s%s", e.Request.URL.Scheme, e.Request.URL.Hostname(), link)
			tmpDir = filepath.Join(outputDir, name)

			//fmt.Println("tmpDir :", tmpDir)
			//fmt.Println("link :", linkURL)
			if link != "" {
				detailLink.Visit(linkURL)
			}
		}

	})

	detailLink.OnHTML(".mdui-row > ul > li > a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		name := e.ChildText("span")

		if strings.HasSuffix(link, ".mp4") {
			fmt.Printf("\t%s\n", name)
			downLoadURL := fmt.Sprintf("%s://%s%s", e.Request.URL.Scheme, e.Request.URL.Hostname(), link)
			//fmt.Println("link :", downLoadURL)
			err := videoDLWorker(filepath.Join(tmpDir, name), downLoadURL)
			if err != nil {
				log.Println(err)
			}
		}
	})

	c.Visit(targetUrl)

}
