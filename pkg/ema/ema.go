package ema

import "golang.org/x/exp/constraints"

// Number is an interface for a numeric value.
type Number interface {
	constraints.Float | constraints.Signed | constraints.Unsigned
}

// EMA store the context of an exponential moving average.
type EMA[T Number] struct {
	alpha float64
	value float64
	count int
}

// NewEMA returns a new EMA struct, given an alpha value.
func NewEMA[T Number](alpha float64) *EMA[T] {
	return &EMA[T]{
		alpha: alpha,
		value: 0,
		count: 0,
	}
}

// NewDefaultEMA returns a new EMA struct. The alpha value is automatically
// calculated from n, which represents the width of the averaging window.
func NewDefaultEMA[T Number](n int) *EMA[T] {
	alpha := 2.0 / (1.0 + float64(n))
	return NewEMA[T](alpha)
}

// Average returns the current average value.
func (e *EMA[T]) Average() T {
	return T(e.value)
}

// Update updates the EMA value.
func (e *EMA[T]) Update(x T) {
	e.value = e.alpha*float64(x) + (1-e.alpha)*e.value
	e.count++
}

// Count returns the number of added samples.
func (e *EMA[T]) Count() int {
	return e.count
}
