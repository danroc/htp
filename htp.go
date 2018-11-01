package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"
)

func main() {
	n := flag.Uint("n", 12, "Number of requests")
	h := flag.String("u", "https://www.google.com", "Host URL")
	flag.Parse()

	lo, hi := math.Inf(-1), math.Inf(+1)
	for i := uint(0); i < *n; i++ {
		t0 := time.Now()
		resp, err := http.Head(*h)
		t1 := time.Now()
		if err != nil {
			log.Fatal(err)
		}

		resp.Body.Close()
		date := resp.Header.Get("Date")
		t2, err := time.Parse(time.RFC1123, date)
		if err != nil {
			log.Fatal(err)
		}

		u2 := t2.UnixNano()
		d0 := float64(t0.UnixNano()-u2) / 1e9
		d1 := float64(t1.UnixNano()-u2) / 1e9

		lo = math.Max(lo, d0-1)
		hi = math.Min(hi, d1)
	}
	offset := (hi + lo) / 2
	margin := (hi - lo) / 2
	now := time.Now().
		Add(time.Duration(-offset * 1e9)).
		Format(time.RFC3339Nano)

	fmt.Printf("%s\n", now)
	fmt.Printf("offset: %.3f (± %.3f) sec.\n", offset, margin)
}
