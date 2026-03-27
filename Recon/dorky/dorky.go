package main

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jpillora/opts"
)

type Config struct {
	Target string `opts:"short=t, help=Name to dork for"`
}

func getFiles() []string {
	dirPath := "/usr/share/dorky/queries/"
	dir, _ := os.ReadDir(dirPath)
	var engines []string

	for file := range dir {
		fileName := string(dir[file].Name())
		filePath := filepath.Join(dirPath, fileName)
		engines = append(engines, filePath)
	}

	return engines
}

func randUA() string {
	ua := []string{
		"Mozilla/5.0 (Linux; Android 14; Pixel 9 Pro Build/AD1A.240418.003; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/124.0.6367.54 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; 23129RAA4G Build/TKQ1.221114.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/116.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (iPhone17,2; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Resorts/4.5.2",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15",
	}

	i := rand.IntN(len(ua) - 1)
	chosenUA := ua[i]
	return chosenUA
}

func setTarget(queries []string) {}

func dork(queries []string, engine string) (results []string, err error) {
	var baseURL string

	switch filepath.Base(engine) {
	case "google":
		baseURL = "https://www.google.com/search?q="

	case "shodan":
	case "bing":
	case "github":
	}
	for n, query := range queries {

		fmt.Printf("Working on %s dorks, doing %d of %d\n", filepath.Base(engine), n, len(queries))

		encQuery := url.QueryEscape(query)
		req, err := http.NewRequest("GET", baseURL+encQuery, nil)
		ua := randUA()

		req.Header.Set("User-Agent", ua)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return []string{}, fmt.Errorf("Error during %s dorks, with query %s:%w", engine, query, err)
		}
		defer res.Body.Close()

		size, err := io.Copy(io.Discard, res.Body)

		if res.StatusCode == 200 && size >= 1024 && err == nil /*Placeholder*/ {
			fmt.Printf("Query %d was succesful (dork: %s)\n", n, query)
			fmt.Printf("Response size to set baseline (TESTING ONLY): %v\n", size)
			fmt.Print(res.ContentLength)
			results = append(results, query)
		} else {
			return []string{}, err
		}
		time.Sleep(5 * time.Second)
		time.Sleep(time.Duration(rand.Int64N(10)) * time.Second)
	}
	return results, nil
}

func main() {
	c := Config{}
	opts.Parse(&c)

	start := time.Now()
	fmt.Printf("Starting dorky at %v", start)

	engines := getFiles()
	fmt.Printf("engines format %T and value %v\n\n", engines, engines)

	for current := range engines {
		engine := engines[current]
		queriesBytes, err := os.ReadFile(engine)
		if err != nil {
			return
		}
		re := regexp.MustCompile("TARGET")
		setQueries := re.ReplaceAllString(string(queriesBytes), c.Target)
		queries := strings.Split(setQueries, "\n")
		dork(queries, engine)
	}

	elapsed := time.Since(start)
	fmt.Printf("Dorky has finished searching in %v", elapsed)
}
