package main

import (
	"fmt"
	"net/http"
	"time"
)

type Target struct {
	Name string
	URL  string
}

type Result struct {
	Target     Target
	StatusCode int
	Up         bool
	Latency    time.Duration
	Err        error
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
			Target:  t,
			Latency: latency,
			Err:     fmt.Errorf("request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	return Result{
		Target:     t,
		StatusCode: resp.StatusCode,
		Up:         isHealthy(resp.StatusCode),
		Latency:    latency,
	}
}

func main() {
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
		{
			Name: "dns-fail",
			URL:  "https://nope-xyz.invalid",
		},
	}

	var results []Result

	for _, t := range targets {
		results = append(results, check(t))
	}

	counts := map[bool]int{}
	for _, r := range results {
		counts[r.Up]++
		if r.Err != nil {
			fmt.Printf("%-4s %-8s   -  %4dms  %s   (%v)\n",
				"DOWN", r.Target.Name, r.Latency.Milliseconds(), r.Target.URL, r.Err)
			continue
		}
		status := "DOWN"
		if r.Up {
			status = "UP"
		}
		fmt.Printf("%-4s %-8s %3d  %4dms  %s\n",
			status, r.Target.Name, r.StatusCode, r.Latency.Milliseconds(), r.Target.URL)

	}
	fmt.Printf("%d up, %d down (%d total)\n", counts[true], counts[false], len(results))

}
