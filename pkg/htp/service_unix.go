//go:build darwin || linux

package htp

import "golang.org/x/sys/unix"

// SyncSystem synchronizes the system clock using a SyncModel.
func SyncSystem(model *SyncModel) error {
	tv := unix.NsecToTimeval(model.Now().UnixNano())
	return unix.Settimeofday(&tv)
}
