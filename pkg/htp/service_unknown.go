//go:build !(darwin || (linux && (amd64 || arm64 || 386)) || windows)

package htp

import (
	"fmt"
	"runtime"
)

func SyncSystem(model *SyncModel) error {
	return fmt.Errorf("system not supported: %s", runtime.GOOS)
}
