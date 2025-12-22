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

// NanoSec represents a time or duration in nanoseconds.
type NanoSec int64

// Sec converts a nanosecond value to seconds.
func (ns NanoSec) Sec() float64 {
	return float64(ns) / float64(time.Second)
}

// SyncRound stores the timestamps of a synchronization round.
type SyncRound struct {
	Send    NanoSec
	Remote  NanoSec
	Receive NanoSec
}

// SyncModel stores the state of the model used in a HTP synchronization.
type SyncModel struct {
	lower NanoSec
	upper NanoSec
	rtt   *ema.EMA[NanoSec]
}

// RTT returns the round-trip time of a SyncRound.
func (r *SyncRound) RTT() NanoSec {
	return r.Receive - r.Send
}

// NewSyncModel returns a new SyncModel.
func NewSyncModel() *SyncModel {
	return &SyncModel{
		lower: math.MinInt64,
		upper: math.MaxInt64,
		rtt:   ema.NewDefaultEMA[NanoSec](samples),
	}
}

// Update updates the model given a new synchronization round.
func (s *SyncModel) Update(round *SyncRound) error {
	var (
		t0 = round.Send
		t1 = round.Remote
		t2 = round.Receive
	)

	s.rtt.Update(round.RTT())
	s.lower = max(s.lower, t0-t1-second)
	s.upper = min(s.upper, t2-t1)

	if s.lower > s.upper {
		return errors.New("local or remote clock changed")
	}
	return nil
}

// Offset returns the current estimate of the offset between local and remote
// clocks.
func (s *SyncModel) Offset() NanoSec {
	return (s.upper + s.lower) / 2
}

// Offset returns the current synchronization error margin.
func (s *SyncModel) Margin() NanoSec {
	return (s.upper - s.lower) / 2
}

// RTT returns the current average of the round-trip-time.
func (s *SyncModel) RTT() NanoSec {
	return s.rtt.Average()
}

// Count returns the number of rounds used to update the model.
func (s *SyncModel) Count() int {
	return s.rtt.Count()
}

// Delay returns the delay from now that we should wait before sending the next
// HTTP request.
func (s *SyncModel) Delay(now NanoSec) time.Duration {
	if s.Count() == 0 {
		return time.Duration(0)
	}
	return time.Duration(secMod(s.Offset() - s.RTT()/2 - now))
}

// Sleep waits (calls time.Sleep()) until it's time to send the next HTTP
// request.
func (s *SyncModel) Sleep() {
	now := NanoSec(time.Now().UnixNano())
	time.Sleep(s.Delay(now))
}

// Now returns the estimate of the current time of the remote host.
func (s *SyncModel) Now() time.Time {
	return time.Now().Add(time.Duration(-s.Offset()))
}

// secMod returns the number of nanoseconds past a round second value.
func secMod(x NanoSec) NanoSec {
	y := x % second
	if y >= 0 {
		return y
	}
	return second + y
}
