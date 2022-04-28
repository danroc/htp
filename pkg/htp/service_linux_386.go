//go:build linux && 386

package htp

import "golang.org/x/sys/unix"

func SyncSystem(model *SyncModel) error {
	now := model.Now()
	tv := &unix.Timeval{
		Sec:  int32(now.Unix()),
		Usec: int32(now.UnixMicro() % 1_000_000),
	}
	return unix.Settimeofday(tv)
}
