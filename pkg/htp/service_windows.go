//go:build windows

package htp

import (
	"fmt"
	"os/exec"
)

// SyncSystem synchronizes the system clock using a SyncModel.
func SyncSystem(model *SyncModel) error {
	arg := fmt.Sprintf(
		"Set-Date -Adjust $([TimeSpan]::FromSeconds(%+.3f))",
		-model.Offset().Sec(),
	)
	return exec.Command("powershell", "-Command", arg).Run()
}
