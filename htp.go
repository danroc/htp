package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/danroc/htp/pkg/htp"
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
	if !strings.Contains(opts.host, "://") {
		opts.host = "https://" + opts.host
	}

	logger := log.New(os.Stderr, "", 0)
	model := htp.NewSyncModel()
	client, err := htp.NewSyncClient(opts.host, 10*time.Second)
	if err != nil {
		logger.Fatal("Cannot create client: ", err)
	}

	options := &htp.SyncOptions{
		Count: int(opts.count),
		Trace: func(round *htp.SyncRound) {
			if opts.verbose {
				margin := model.Margin()
				logger.Printf("offset: %+.3f (±%.3f) seconds\n",
					toSec(model.Offset()), toSec(margin))
			}
		},
	}
	if err = htp.Sync(client, model, options); err != nil {
		logger.Fatal("Cannot sync clock: ", err)
	}

	if opts.date {
		now := time.Now().Add(time.Duration(-model.Offset()))
		fmt.Printf("%s\n", now.Format(opts.format))
	} else {
		fmt.Printf("%+.3f\n", toSec(-model.Offset()))
	}

	if opts.sync {
		if err := syncSystem(model.Offset()); err != nil {
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
