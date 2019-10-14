package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	inputFilename = "ids.txt"
	proxyStr      = "http://177.185.240.241:80"
	URLRoot       = "http://www3.mte.gov.br/sistemas/mediador/Resumo/ResumoVisualizar?NrSolicitacao="
	workers       = 6
)

func main() {

	idPool := make(chan string, workers)
	wg := &sync.WaitGroup{}

	file, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idPool {
				if err := makeRequest(id); err != nil {
					log.Printf("Error when processing id: \"%s\": %v", id, err)
				}
			}
		}()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := scanner.Text()
		idPool <- id
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}

func makeRequest(id string) error {
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return err
	}
	idWithDash := strings.Replace(id, "/", "-", -1)

	// creating the URL to be loaded through the proxy
	URLStr := fmt.Sprintf("%s%s", URLRoot, id)
	URL, err := url.Parse(URLStr)
	if err != nil {
		return err
	}

	// adding the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// adding the Transport object to the http Client
	client := &http.Client{
		Transport: transport,
	}

	// generating the HTTP GET request
	request, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %v", err)
	}

	// calling the URL
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("client.Do: %v", err)
	}
	defer response.Body.Close()

	// getting the response
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %v", err)
	}

	file, err := os.Create(path.Join("DATUMS", idWithDash))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("file.Write: %v", err)
	}
	return nil
}
