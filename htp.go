package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

// Constants from "time" package but with the int64 type.
const (
	Nanosecond  int64 = 1
	Microsecond       = 1000 * Nanosecond
	Millisecond       = 1000 * Microsecond
	Second            = 1000 * Millisecond
)

func main() {
	host := flag.String("u", "https://google.com", "Host URL")
	count := flag.Uint("n", 8, "Number of requests")
	quiet := flag.Bool("q", false, "Do not output time offset")
	flag.Parse()

	logger := log.New(os.Stderr, "", 0)
	http.DefaultClient.CheckRedirect = noRedirect

	var (
		offset int64
		sleep  int64 = 0
		lo     int64 = math.MinInt64
		hi     int64 = math.MaxInt64
	)
	for i := uint(0); i < *count; i++ {
		time.Sleep(time.Duration(sleep))

		t0 := time.Now().UnixNano()
		resp, err := http.Head(*host)
		t1 := time.Now().UnixNano()

		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()

		dateStr := resp.Header.Get("Date")
		date, err := time.Parse(time.RFC1123, dateStr)
		if err != nil {
			logger.Fatalf("Invalid HTTP response date: \"%s\"", dateStr)
		}
		t2 := date.UnixNano()

		lo = max(lo, t0-t2-Second)
		hi = min(hi, t1-t2)
		offset = (hi + lo) / 2
		if !*quiet {
			margin := (hi - lo) / 2
			logger.Printf("offset: %+.3f (±%.3f) seconds\n",
				float64(offset)/float64(Second),
				float64(margin)/float64(Second))
		}

		sleep = offset - (t1-t0)/2 - t1%Second
		for sleep < 0 {
			sleep += Second
		}
		for sleep > Second {
			sleep -= Second
		}
	}

	now := time.Now().Add(time.Duration(-offset))
	fmt.Printf("%s\n", now.Format(time.RFC3339Nano))
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
