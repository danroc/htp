package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptrace"
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

const UnixDateMilli = "Mon Jan _2 15:04:05.000 MST 2006"

func main() {
	host := flag.String("u", "https://google.com", "Host URL")
	count := flag.Uint("n", 8, "Number of requests")
	quiet := flag.Bool("q", false, "Do not output time offsets")
	layout := flag.String("f", UnixDateMilli, "Time format layout")
	flag.Parse()

	logger := log.New(os.Stderr, "", 0)
	client := http.DefaultClient
	client.CheckRedirect = noRedirect
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("HEAD", *host, nil)
	if err != nil {
		logger.Fatal("Invalid HTTP request: ", err)
	}
	req.Header.Add("Cache-Control", "no-cache")

	var (
		t0, t1 int64
		offset int64
		sleep  int64 = 0
		lo     int64 = math.MinInt64
		hi     int64 = math.MaxInt64
	)

	ctx := httptrace.WithClientTrace(req.Context(),
		&httptrace.ClientTrace{
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				t0 = time.Now().UnixNano()
			},
			GotFirstResponseByte: func() {
				t1 = time.Now().UnixNano()
			},
		},
	)

	for i := uint(0); i < *count; i++ {
		time.Sleep(time.Duration(sleep))

		resp, err := client.Do(req.WithContext(ctx))
		if err != nil {
			logger.Fatal("Invalid HTTP response: ", err)
		}
		resp.Body.Close()

		dateStr := resp.Header.Get("Date")
		date, err := time.Parse(time.RFC1123, dateStr)
		if err != nil {
			logger.Fatal("Invalid HTTP response date: ", err)
		}
		t2 := date.UnixNano()

		lo = max(lo, t0-t2-Second)
		hi = min(hi, t1-t2)
		if hi < lo {
			logger.Fatal("Cannot synchronize clocks: " +
				"Local or remote clock changed during synchronization")
		}

		offset = (hi + lo) / 2
		if !*quiet {
			margin := (hi - lo) / 2
			logger.Printf("offset: %+.3f (±%.3f) seconds\n",
				float64(offset)/float64(Second),
				float64(margin)/float64(Second))
		}

		sleep = offset - (t1-t0)/2 - t1%Second
		sleep = mod(sleep, Second)
	}

	now := time.Now().Add(time.Duration(-offset))
	fmt.Printf("%s\n", now.Format(*layout))
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

func mod(x, m int64) int64 {
	y := x % m
	if y >= 0 {
		return y
	}
	return m + y
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
