package htp

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

const (
	// isoFormat   = "2006-01-02T15:04:05.000Z07:00"
	// unixFormat  = "Mon Jan _2 15:04:05.000 MST 2006"
	macosFormat = "0102150406.05"
)

type SyncTrace struct {
	Before func(i int) bool
	After  func(i int, round *SyncRound) bool
}

type SyncOptions struct {
	Count int
	Trace *SyncTrace
}

func Sync(client *SyncClient, model *SyncModel, trace *SyncTrace) error {
	for i := 0; ; i++ {
		if !trace.Before(i) {
			break
		}

		model.Sleep()

		round, err := client.Round()
		if err != nil {
			return err
		}

		if err := model.Update(round); err != nil {
			return err
		}

		if !trace.After(i, round) {
			break
		}
	}

	return nil
}

func SyncSystem(model *SyncModel) error {
	switch runtime.GOOS {
	case "windows":
		arg := fmt.Sprintf("Set-Date -Adjust $([TimeSpan]::FromSeconds(%+.3f))", -model.Offset().Sec())
		return exec.Command("powershell", "-Command", arg).Run()

	case "linux":
		arg := fmt.Sprintf("%+.3f seconds", -model.Offset().Sec())
		return exec.Command("date", "-s", arg).Run()

	case "darwin":
		arg := model.Now().Add(time.Second).Format(macosFormat)
		sleep := time.Duration(int(time.Second) - model.Now().Nanosecond())
		time.Sleep(sleep)
		return exec.Command("date", arg).Run()

	default:
		return fmt.Errorf("system not supported: %s", runtime.GOOS)
	}
}
