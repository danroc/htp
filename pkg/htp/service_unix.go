//go:build darwin || linux

package htp

import "golang.org/x/sys/unix"

func syncSystem(model *SyncModel) error {
	tv := unix.NsecToTimeval(model.Now().UnixNano())
	return unix.Settimeofday(&tv)
}
