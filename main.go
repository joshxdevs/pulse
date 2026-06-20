package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Target struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Result struct {
	Target     Target        `json:"target"`
	StatusCode int           `json:"status_code"`
	Up         bool          `json:"up"`
	Latency    time.Duration `json:"-"`
	LatencyMS  int64         `json:"latency_ms"`
	Err        error         `json:"-"`
	Error      string        `json:"error,omitempty"`
}

func (r Result) Status() string {
	if r.Up {
		return "UP"
	}
	return "DOWN"
}

func (r Result) String() string {
	if r.Err != nil {
		return fmt.Sprintf("%-4s %-8s   -  %4dms  %s   (%v)",
			r.Status(), r.Target.Name, r.Latency.Milliseconds(), r.Target.URL, r.Err)
	}
	return fmt.Sprintf("%-4s %-8s %3d  %4dms  %s",
		r.Status(), r.Target.Name, r.StatusCode, r.Latency.Milliseconds(), r.Target.URL)
}

type Reporter interface {
	Report(results []Result) error
}

type textReporter struct{}

func (textReporter) Report(results []Result) error {
	up, down := 0, 0
	for _, r := range results {
		fmt.Println(r)
		if r.Up {
			up++
		} else {
			down++
		}
	}
	fmt.Printf("%d up, %d down (%d total)\n", up, down, len(results))
	return nil
}

type JSONReporter struct{}

func (JSONReporter) Report(results []Result) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func isHealthy(code int) bool {
	return code >= 200 && code <= 399
}

func check(t Target) Result {
	start := time.Now()
	resp, err := http.Get(t.URL)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Target:    t,
			Latency:   latency,
			LatencyMS: latency.Milliseconds(),
			Err:       err,
			Error:     err.Error(),
		}
	}
	defer resp.Body.Close()

	return Result{
		Target:     t,
		StatusCode: resp.StatusCode,
		Up:         isHealthy(resp.StatusCode),
		Latency:    latency,
		LatencyMS:  latency.Milliseconds(),
	}
}

func main() {
	format := flag.String("format", "text", "output format: text|json")
	flag.Parse()

	targets := []Target{
		{
			Name: "google",
			URL:  "https://google.com",
		},
		{
			Name: "github",
			URL:  "https://github.com",
		},
		{
			Name: "broken",
			URL:  "https://httpstat.us/500",
		},
	}

	start := time.Now()

	results := make([]Result, len(targets))
	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = check(t)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	var reporter Reporter = textReporter{}
	if *format == "json" {
		reporter = JSONReporter{}
	}

	if err := reporter.Report(results); err != nil {
		fmt.Fprintln(os.Stderr, "report error:", err)
		os.Exit(1)
	}
	fmt.Printf("checked %d targets in %s\n", len(targets), elapsed.Round(time.Millisecond))
}
