package main

import (
	"context"
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

func check(ctx context.Context, t Target) Result {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return Result{Target: t, Err: err, Error: err.Error()}
	}
	resp, err := http.DefaultClient.Do(req)
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

func runChecks(ctx context.Context, targets []Target, timeout time.Duration, workers int) []Result {
	jobs := make(chan Target, len(targets))
	resultsCh := make(chan Result, len(targets))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				reqctx, cancel := context.WithTimeout(ctx, timeout)
				resultsCh <- check(reqctx, t)
				cancel()
			}
		}()
	}

	for _, t := range targets {
		jobs <- t
	}
	close(jobs)
	go func() { wg.Wait(); close(resultsCh) }()

	var results []Result
	for r := range resultsCh {
		results = append(results, r)
	}
	return results
}

func main() {
	format := flag.String("format", "text", "output format: text|json")
	timeout := flag.Duration("timeout", 5*time.Second, "per-request timeout")
	concurrency := flag.Int("concurrency", 8, "max parallel checks")
	flag.Parse()

	targets := []Target{
		{Name: "google", URL: "https://google.com"},
		{Name: "slow", URL: "https://httpstat.us/200?sleep=10000"},
	}

	start := time.Now()
	results := runChecks(context.Background(), targets, *timeout, *concurrency)
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
