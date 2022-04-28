//go:build linux && (amd64 || arm64)

package htp

import "golang.org/x/sys/unix"

func SyncSystem(model *SyncModel) error {
	now := model.Now()
	tv := &unix.Timeval{
		Sec:  now.Unix(),
		Usec: now.UnixMicro() % 1_000_000,
	}
	return unix.Settimeofday(tv)
}
