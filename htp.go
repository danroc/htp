package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"
)

// Constants from "time" package but with the int64 type.
const (
	Nanosecond  int64 = 1
	Microsecond       = 1000 * Nanosecond
	Millisecond       = 1000 * Microsecond
	Second            = 1000 * Millisecond
)

const unixDateMilli = "02 Jan 2006 15:04:05.000 MST"

func main() {
	host := flag.String("u", "https://google.com", "Host URL")
	count := flag.Uint("n", 8, "Number of requests")
	verbose := flag.Bool("v", false, "Verbose mode (display offsets)")
	showDate := flag.Bool("d", false, "Display date and time instead of offset")
	dateLayout := flag.String("f", unixDateMilli, "Time format layout")
	flag.Parse()

	if strings.Index(*host, "://") == -1 {
		*host = "http://" + *host
	}

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
		sleep  int64
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
		if *verbose {
			margin := (hi - lo) / 2
			logger.Printf("offset: %+.3f (±%.3f) seconds\n",
				toSec(offset), toSec(margin))
		}

		sleep = offset - (t1-t0)/2 - t1%Second
		sleep = mod(sleep, Second)
	}

	if *showDate {
		now := time.Now().Add(time.Duration(-offset))
		fmt.Printf("%s\n", now.Format(*dateLayout))
	} else {
		fmt.Printf("%+.3f seconds\n", toSec(-offset))
	}
}

func toSec(t int64) float64 {
	return float64(t) / float64(Second)
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
