package ema

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Float | constraints.Signed | constraints.Unsigned
}

type EMA[T Number] struct {
	alpha float64
	value float64
	count int
}

func NewEMA[T Number](alpha float64) *EMA[T] {
	return &EMA[T]{
		alpha: alpha,
		value: 0,
		count: 0,
	}
}

func NewDefaultEMA[T Number](n int) *EMA[T] {
	alpha := 2.0 / (1.0 + float64(n))
	return NewEMA[T](alpha)
}

func (e *EMA[T]) Average() T {
	return T(e.value)
}

func (e *EMA[T]) Update(x T) {
	e.value = e.alpha*float64(x) + (1-e.alpha)*e.value
	e.count++
}

func (e *EMA[T]) Count() int {
	return e.count
}
