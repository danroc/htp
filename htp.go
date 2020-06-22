package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	// isoFormat   = "2006-01-02T15:04:05.000Z07:00"
	// unixFormat  = "Mon Jan _2 15:04:05.000 MST 2006"
	macosFormat = "0102150406.05"
	second      = int64(time.Second)
)

type options struct {
	host    string
	count   uint
	verbose bool
	date    bool
	sync    bool
	format  string
}

func main() {
	opts := parseArgs()
	if strings.Index(opts.host, "://") == -1 {
		opts.host = "https://" + opts.host
	}

	logger := log.New(os.Stderr, "", 0)
	client := &http.Client{}
	client.Timeout = 10 * time.Second
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest("HEAD", opts.host, nil)
	if err != nil {
		logger.Fatal("Invalid HTTP request: ", err)
	}

	var (
		t0, t2 int64
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
				t2 = time.Now().UnixNano()
			},
		},
	)

	for i := uint(0); i < opts.count; i++ {
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
		t1 := date.UnixNano()

		lo = max(lo, t0-t1-second)
		hi = min(hi, t2-t1)
		if hi < lo {
			logger.Fatal("Cannot synchronize clocks: " +
				"local or remote clock changed during synchronization")
		}

		offset = (hi + lo) / 2
		if opts.verbose {
			margin := (hi - lo) / 2
			logger.Printf("offset: %+.3f (±%.3f) seconds\n",
				toSec(offset), toSec(margin))
		}

		sleep = offset - (t2-t0)/2 - t2
		sleep = mod(sleep, second)
	}

	if opts.date {
		now := time.Now().Add(time.Duration(-offset))
		fmt.Printf("%s\n", now.Format(opts.format))
	} else {
		fmt.Printf("%+.3f\n", toSec(-offset))
	}

	if opts.sync {
		if err := syncSystem(offset); err != nil {
			logger.Fatal("Cannot set system clock: ", err)
		}
	}
}

func syncSystem(offset int64) error {
	switch runtime.GOOS {
	case "windows":
		arg := fmt.Sprintf("Set-Date -Adjust $([TimeSpan]::FromSeconds(%+.3f))", toSec(-offset))
		return exec.Command("powershell", "-Command", arg).Run()
	case "linux":
		arg := fmt.Sprintf("%+.3f seconds", toSec(-offset))
		return exec.Command("date", "-s", arg).Run()
	case "darwin":
		now := time.Now().Add(time.Duration(-offset))
		arg := now.Add(time.Second).Format(macosFormat)
		sleep := time.Duration(int(time.Second) - now.Nanosecond())
		time.Sleep(sleep)
		return exec.Command("date", arg).Run()
	default:
		return fmt.Errorf("system not supported: %s", runtime.GOOS)
	}
}

func parseArgs() *options {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"HTP - Date and time from HTTP headers\n\nUsage:\n")
		flag.PrintDefaults()
	}

	opts := options{}
	flag.StringVar(&opts.host, "u", "https://www.google.com", "Host URL")
	flag.UintVar(&opts.count, "n", 8, "Number of requests")
	flag.BoolVar(&opts.verbose, "v", false, "Show offsets during synchronization")
	flag.BoolVar(&opts.date, "d", false, "Display date and time instead of offset")
	flag.StringVar(&opts.format, "f", time.UnixDate, "Date and time format")
	flag.BoolVar(&opts.sync, "s", false, "Synchronize system time")
	flag.Parse()

	return &opts
}

func toSec(t int64) float64 {
	return float64(t) / float64(second)
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
