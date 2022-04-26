package htp

import (
	"errors"
	"math"
	"time"

	"github.com/danroc/htp/pkg/ema"
)

const (
	second  = NanoSec(time.Second) // Number of nanoseconds in one second
	samples = 3                    // Number of samples in our moving average
)

type NanoSec int64

func (ns NanoSec) Sec() float64 {
	return float64(ns) / float64(time.Second)
}

type SyncRound struct {
	Send    NanoSec
	Remote  NanoSec
	Receive NanoSec
}

type SyncModel struct {
	count int
	lower NanoSec
	upper NanoSec
	rtt   *ema.EMA[NanoSec]
}

func NewSyncModel() *SyncModel {
	return &SyncModel{
		count: 0,
		lower: math.MinInt64,
		upper: math.MaxInt64,
		rtt:   ema.NewDefaultEMA[NanoSec](samples),
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

func (s *SyncModel) Offset() NanoSec {
	return (s.upper + s.lower) / 2
}

func (s *SyncModel) Margin() NanoSec {
	return (s.upper - s.lower) / 2
}

func (s *SyncModel) RTT() NanoSec {
	return s.rtt.Average()
}

func (s *SyncModel) Count() int {
	return s.count
}

func (s *SyncModel) Delay(now NanoSec) time.Duration {
	if s.count == 0 {
		return time.Duration(0)
	}
	return time.Duration(secMod(s.Offset() - s.RTT()/2 - now))
}

func (s *SyncModel) Sleep() {
	now := NanoSec(time.Now().UnixNano())
	time.Sleep(s.Delay(now))
}

func min(a, b NanoSec) NanoSec {
	if a < b {
		return a
	}
	return b
}

func max(a, b NanoSec) NanoSec {
	if a > b {
		return a
	}
	return b
}

func secMod(x NanoSec) NanoSec {
	y := x % second
	if y >= 0 {
		return y
	}
	return second + y
}
