package main

import (
	"flag"
	"fmt"
	"net/http"
)

type Target struct {
	Name string
	URL  string
}

func check(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func isHealthy(code int) bool {
	return code >= 200 && code <= 399
}

func main() {
	url := flag.String("url", "https://example.com", "the URL to check")
	name := flag.String("name", "unnamed", "a label for the target")
	flag.Parse()

	target := Target{Name: *name, URL: *url}

	code, err := check(target.URL)
	if err != nil {
		fmt.Printf("%-4s %s (%v) %s\n", "DOWN", target.Name, err, target.URL)
		return
	}

	status := "DOWN"
	if isHealthy(code) {
		status = "UP"
	}
	fmt.Printf("%-4s %s (%d) %s\n", status, target.Name, code, target.URL)
}
