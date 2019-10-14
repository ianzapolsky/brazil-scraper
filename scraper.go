// Every Golang program must have a main package. Library packages can also be
// specified, which can then be imported by other programs.
package main

// Import standard library packages
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

// Consts are variables that cannot be mutated in any way after initial
// declaration. This is useful for variables that we do not expect to change
// over the lifetime of our program.
const (
	inputFilename = "ids.txt"
	dataDir       = "DATUMS"
	proxyStr      = "http://177.185.240.241:80"
	URLRoot       = "http://www3.mte.gov.br/sistemas/mediador/Resumo/ResumoVisualizar?NrSolicitacao="
	// This variable controls how many parallel processes to run at once
	workers = 10
)

// Every Golang program must have a main() function. It is what the compiler
// will run
func main() {

	// The next several lines is an implementation of a simple worker pool.

	// This is a channel, a special structure in Golang. Basically, it is a pipe
	// that we can write data to and read data from simultaneously, without
	// having to worry about synchronization (e.g. locking).
	// More details here: https://tour.golang.org/concurrency/2
	idChannel := make(chan string, workers)
	// This is a WaitGroup. We use this to ensure that our program does not exit
	// before all the workers we are about to start finish execution.
	wg := &sync.WaitGroup{}
	// In this for loop, we will create one new goroutine for each worker
	for i := 0; i < workers; i++ {
		// First, add 1 to our WaitGroup
		wg.Add(1)
		// This is an inline declaration of a function, that we invoke with "go"
		// which causes the function to run as a separate goroutine. Execution
		// of this function therefore is not blocking, and our main() function
		// continues on immediately to do other stuff.
		go func() {
			defer wg.Done()
			// Each worker reads IDs out of the idChannel and then calls the
			// function makeRequest on them. These workers will run in parallel.
			for id := range idChannel {
				if err := makeRequest(id); err != nil {
					log.Printf("Error when processing id: \"%s\": %v", id, err)
				}
			}
		}()
	}

	// This opens the inputFile
	file, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// A Scanner allows us to read a file line by line, instead of buffering
	// it all into memory first
	scanner := bufio.NewScanner(file)
	// This loop exits when there are no more lines to read from our file
	for scanner.Scan() {
		// Each line of our input file contains one ID, e.g. "MR003081/2008"
		id := scanner.Text()
		// In this line, we put the ID into our channel. The worker goroutines
		// set up above are simultaneously reading IDs from this channel and
		// calling makeRequest() on each ID
		idChannel <- id
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// Here, we call Wait on our WaitGroup to ensure that we do not exit before
	// all of our worker goroutines have finished.
	wg.Wait()
}

func makeRequest(id string) error {
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return err
	}

	// Create the URL to be loaded through the proxy
	URLStr := fmt.Sprintf("%s%s", URLRoot, id)
	URL, err := url.Parse(URLStr)
	if err != nil {
		return err
	}

	// Add the proxy settings to the Transport object
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Adding the Transport object to the http Client
	client := &http.Client{
		Transport: transport,
	}

	// Generate the HTTP GET request
	request, err := http.NewRequest("GET", URL.String(), nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest: %v", err)
	}

	// Make the request
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("client.Do: %v", err)
	}
	defer response.Body.Close()

	// Read the response
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll: %v", err)
	}

	// Create an output file
	idWithDash := strings.Replace(id, "/", "-", -1)
	file, err := os.Create(path.Join(dataDir, idWithDash))
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the data to the output file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("file.Write: %v", err)
	}
	return nil
}
