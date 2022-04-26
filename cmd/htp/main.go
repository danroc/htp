package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/danroc/htp/pkg/htp"
	"github.com/spf13/cobra"
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
		Use:          "htp",
		Short:        "HTP - Date and time from HTTP headers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.Contains(host, "://") {
				host = "https://" + host
			}

			model := htp.NewSyncModel()
			client, err := htp.NewSyncClient(host, time.Duration(timeout)*time.Second)
			if err != nil {
				return nil
			}

			trace := &htp.SyncTrace{
				Before: func(i int) bool { return i < count },
				After: func(i int, round *htp.SyncRound) bool {
					if !silent {
						fmt.Fprintf(os.Stderr, "(%d/%d) offset: %+.3f (±%.3f) seconds\n",
							i+1, count, model.Offset().Sec(), model.Margin().Sec())
					}
					return true
				},
			}

			if err := htp.Sync(client, model, trace); err != nil {
				return err
			}

			if date {
				fmt.Printf("%s\n", model.Now().Format(format))
			} else {
				fmt.Printf("%+.3f\n", -model.Offset().Sec())
			}

			if sync {
				if err := htp.SyncSystem(model); err != nil {
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
