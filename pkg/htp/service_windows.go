//go:build windows

package htp

import (
	"fmt"
	"os/exec"
)

func syncSystem(model *SyncModel) error {
	arg := fmt.Sprintf("Set-Date -Adjust $([TimeSpan]::FromSeconds(%+.3f))", -model.Offset().Sec())
	return exec.Command("powershell", "-Command", arg).Run()
}
