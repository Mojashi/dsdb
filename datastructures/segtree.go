package datastructures

import "errors"

type SegmentTree struct {
	Data []int
	N    int
}

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
func Min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

func (s *SegmentTree) Init(n int) {
	s.N = n
	m := 1
	for m < s.N {
		m *= 2
	}
	s.Data = make([]int, m*2)
}

func (s SegmentTree) SumInner(l, r, rl, rr, idx int) int {
	l = Max(l, rl)
	r = Min(r, rr)
	if l >= r {
		return 0
	}
	if l == rl && r == rr {
		return s.Data[idx]
	}
	if rr-rl <= 1 {
		return 0
	}

	mid := (rl + rr) / 2
	return s.SumInner(l, Min(r, mid), rl, mid, idx*2) + s.SumInner(Max(l, mid), r, mid, rr, idx*2+1)
}

func (s SegmentTree) Sum(l int, r int) int {
	return s.SumInner(l, r, 0, len(s.Data)/2, 1)
}

func (s SegmentTree) Update(idx int, val int) error {
	m := len(s.Data) / 2
	if idx >= s.N || idx < 0 {
		return errors.New("index out of range")
	}
	idx += m
	s.Data[idx] = val
	idx /= 2
	for idx > 0 {
		s.Data[idx] = s.Data[idx*2] + s.Data[idx*2+1]
		idx /= 2
	}
	return nil
}
