package htp

import (
	"errors"
	"math"
	"time"

	"github.com/danroc/htp/pkg/ema"
)

const (
	second  = int64(time.Second)
	samples = 3
)

type SyncRound struct {
	Send    int64
	Remote  int64
	Receive int64
}

type SyncModel struct {
	count int
	lower int64
	upper int64
	rtt   *ema.EMA[int64]
}

func NewSyncModel() *SyncModel {
	return &SyncModel{
		count: 0,
		lower: math.MinInt64,
		upper: math.MaxInt64,
		rtt:   ema.NewDefaultEMA[int64](samples),
	}
}

func (s *SyncModel) Update(round *SyncRound) error {
	var (
		t0 = round.Send
		t1 = round.Remote
		t2 = round.Receive
	)

	s.rtt.Update(t2 - t0)
	s.lower = max(s.lower, t0-t1-second)
	s.upper = min(s.upper, t2-t1)
	s.count++

	if s.lower > s.upper {
		return errors.New("local or remote clock changed")
	}
	return nil
}

func (s *SyncModel) Offset() int64 {
	return (s.upper + s.lower) / 2
}

func (s *SyncModel) Margin() int64 {
	return (s.upper - s.lower) / 2
}

func (s *SyncModel) RTT() int64 {
	return s.rtt.Average()
}

func (s *SyncModel) Count() int {
	return s.count
}

func (s *SyncModel) Delay(now int64) time.Duration {
	if s.count == 0 {
		return time.Duration(0)
	}
	return time.Duration(secMod(s.Offset() - s.RTT()/2 - now))
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func secMod(x int64) int64 {
	y := x % second
	if y >= 0 {
		return y
	}
	return second + y
}
