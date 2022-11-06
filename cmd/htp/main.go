package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/danroc/htp/pkg/htp"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	var (
		host    string
		silent  bool
		format  string
		sync    bool
		offset  bool
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
					logInfo(
						silent, "(%d/%d) offset: %+.3f (±%.3f) seconds", i+1,
						count, model.Offset().Sec(), model.Margin().Sec())
					return true
				},
			}

			logInfo(silent, "Syncing with %s ...", host)
			if err := htp.Sync(client, model, trace); err != nil {
				return err
			}

			// Always print the result, it can be silenced by the user by
			// redirecting the standard output to /dev/null.
			if offset {
				fmt.Printf("%+.3f\n", -model.Offset().Sec())
			} else {
				fmt.Printf("%s\n", model.Now().Format(format))
			}

			if sync {
				if err := htp.SyncSystem(model); err != nil {
					return fmt.Errorf("cannot set system clock: %w", err)
				}
				logInfo(silent, "System time set")
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "requests", "n", 8, "Number of requests")
	cmd.Flags().IntVarP(&timeout, "timeout", "t", 10, "Timeout in seconds")
	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Do not show offsets")
	cmd.Flags().BoolVarP(&offset, "offset", "o", false, "Show offset instead of date and time")
	cmd.Flags().BoolVarP(&sync, "set", "e", false, "Set system time")
	cmd.Flags().StringVarP(&format, "format", "f", time.UnixDate, "Date and time format")
	cmd.Flags().StringVarP(&host, "url", "u", "https://www.google.com", "Host URL")

	return cmd
}

// All "information" logging must done using this function.
//
// It prints to stderr instead of stdout to allow the user to directly use the
// output of htp in a shell script.
func logInfo(silent bool, format string, args ...interface{}) {
	if !silent {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
