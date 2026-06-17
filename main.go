package main

import (
	"flag"
	"fmt"
)

type Target struct {
	Name string
	URL  string
}

func main() {
	url := flag.String("url", "https://example.com", "the URL to check")
	name := flag.String("name", "unnamed", "a label for the target")
	flag.Parse()

	target := Target{
		Name: *name,
		URL:  *url,
	}

	fmt.Printf("Target %q -> %s\n", target.Name, target.URL)
	fmt.Printf("%+v\n", target)
}
