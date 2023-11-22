package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

func readUrlList(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening URLs file:", err)
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil

}

func downloadFile(urlToDownload string, waitGroup *sync.WaitGroup, semChan chan struct{}) {
	fmt.Println("insideDownload", urlToDownload)
	defer waitGroup.Done()

	parsedUrl, err := url.Parse(urlToDownload)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}
	fileName := "home.html"
	directoriesPath := "htmls"

	if filepath.Base(parsedUrl.Path) != "/" {
		fileName = filepath.Base(parsedUrl.Path) + ".html"
		directoriesPath += filepath.Dir(parsedUrl.Path)
	}

	resp, err := http.Get(urlToDownload)
	if err != nil {
		fmt.Printf("Error fetching %s: %s\n", urlToDownload, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body of %s: %s\n", urlToDownload, err)
		return
	}
	err = os.MkdirAll(directoriesPath, 0755)
	if err != nil {
		fmt.Printf("Error Creating directory for file %s of url:%s with err:%s\n", fileName, urlToDownload, err)
		return
	}

	err = os.WriteFile(directoriesPath+"/"+fileName, body, 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %s\n", fileName, err)
		return
	}

	fmt.Printf("Downloaded %s\n", urlToDownload)
}

func main() {
	urls, err := readUrlList("ListOfAsciiSiteUrls.txt")
	if err != nil {
		fmt.Println("Error Reading URLs:", err)
		return
	}

	maxWorkers := 5
	semaphoreChan := make(chan struct{}, maxWorkers)

	var waitGroup sync.WaitGroup

	for i := range urls {
		waitGroup.Add(1)
		semaphoreChan <- struct{}{}
		go func(currentUrl *string) {
			downloadFile(*currentUrl, &waitGroup, semaphoreChan)
			<-semaphoreChan
		}(&urls[i])
	}

	waitGroup.Wait()
	close(semaphoreChan)

	fmt.Println("All downloads completed.")
}
