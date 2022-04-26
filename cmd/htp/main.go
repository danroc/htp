package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/danroc/htp/pkg/htp"
	"github.com/spf13/cobra"
)

const (
	// isoFormat   = "2006-01-02T15:04:05.000Z07:00"
	// unixFormat  = "Mon Jan _2 15:04:05.000 MST 2006"
	macosFormat = "0102150406.05"
	second      = int64(time.Second)
)

func buildRootCommand() *cobra.Command {
	var (
		host    string
		silent  bool
		format  string
		sync    bool
		date    bool
		count   int
		timeout int
	)

	cmd := &cobra.Command{
		Use:   "htp",
		Short: "HTP - Date and time from HTTP headers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.Contains(host, "://") {
				host = "https://" + host
			}

			client, err := htp.NewSyncClient(host, time.Duration(timeout)*time.Second)
			if err != nil {
				return nil
			}

			model := htp.NewSyncModel()

			options := &htp.SyncOptions{
				Count: int(count),
				Trace: func(i int, round *htp.SyncRound) {
					if !silent {
						fmt.Fprintf(os.Stderr, "(%d/%d) offset: %+.3f (±%.3f) seconds\n",
							i+1, count, toSec(model.Offset()), toSec(model.Margin()))
					}
				},
			}

			if err := htp.Sync(client, model, options); err != nil {
				return err
			}

			if date {
				now := time.Now().Add(time.Duration(-model.Offset()))
				fmt.Printf("%s\n", now.Format(format))
			} else {
				fmt.Printf("%+.3f\n", toSec(-model.Offset()))
			}

			if sync {
				if err := syncSystem(model.Offset()); err != nil {
					return fmt.Errorf("cannot set system clock: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "num-requests", "n", 10, "Number of requests")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Timeout in seconds")
	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Do not show offsets")
	cmd.Flags().BoolVarP(&date, "date", "d", false, "Show date and time instead of offset")
	cmd.Flags().BoolVarP(&sync, "set", "e", false, "Set system time")
	cmd.Flags().StringVarP(&format, "format", "f", time.UnixDate, "Date and time format")
	cmd.Flags().StringVarP(&host, "url", "u", "https://www.google.com", "Host URL")

	return cmd
}

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
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

func toSec(t int64) float64 {
	return float64(t) / float64(second)
}
