package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

func timeDiff(url string) (float64, float64, error) {
	var delta float64
	var t0 time.Time
	trace := &httptrace.ClientTrace{
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			t0 = time.Now()
		},
		GotFirstResponseByte: func() {
			delta = time.Since(t0).Seconds()
		},
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, 0, err
	}
	ctx := httptrace.WithClientTrace(req.Context(), trace)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return 0, 0, err
	}

	date := resp.Header.Get("Date")
	if date == "" {
		return 0, 0, errors.New("Date header is missing")
	}
	t1, err := time.Parse(time.RFC1123, date)
	if err != nil {
		return 0, 0, err
	}

	theta := t1.Sub(t0).Seconds() + 0.5 - delta/2
	return theta, delta, nil
}

func main() {
	n := flag.Uint("n", 10, "Number of requests")
	h := flag.String("u", "https://google.com", "Host URL")
	flag.Parse()

	var sumT, sumD float64
	for i := uint(0); i < *n; i++ {
		theta, delta, err := timeDiff(*h)
		if err != nil {
			log.Fatal(err)
		}
		sumT += theta / delta
		sumD += 1 / delta
	}
	fmt.Printf("offset %.5f sec\n", sumT/sumD)
}
