package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

func main() {

	page, err := http.Get(os.Args[1])

	if err != nil {
		fmt.Printf("got error: %v", err.Error())
	}
	defer page.Body.Close()

	links := getLinks(page.Body)

	wg := sync.WaitGroup{}

	wg.Add(len(links))

	for _, v := range links {
		fmt.Println(v)
		if strings.HasPrefix(v, "//") {
			v = fmt.Sprintf("http:%v", v)
		}
		_, err := url.ParseRequestURI(v)

		if err != nil {
			//panic(err)
			fmt.Printf("Usage: %s [URL]", os.Args[0])
			os.Exit(1)
		}
		go downloadFromUrl(v, &wg)
	}
	wg.Wait()
}

func getLinks(body io.Reader) []string {
	var links []string
	z := html.NewTokenizer(body)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return removeDuplicate(links)
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" && strings.HasSuffix(attr.Val, ".webm") {
						links = append(links, attr.Val)
					}

				}
			}

		}
	}
}

func removeDuplicate(sliceList []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func downloadFromUrl(url string, wg *sync.WaitGroup) {
	defer wg.Done()
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	output, err := os.Create(fmt.Sprintf("%v/%v", os.Args[2], fileName))
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}
