package common

import "sync"
import "sync/atomic"

type ReusableSlice struct {
	slice   []byte
	pool    *sync.Pool
	counter int32
}

func (s *ReusableSlice) AddOwner() {
	atomic.AddInt32(&s.counter, 1)
}

func (s *ReusableSlice) Done() {
	atomic.AddInt32(&s.counter, -1)
	c := atomic.LoadInt32(&s.counter)
	if c == 0 {
		s.pool.Put(s)
	} else if c < 0 {
		panic("incorrect use of ReusableSlice")
	}
}

func (s *ReusableSlice) SlicePtr() *[]byte {
	return &s.slice
}

func (s *ReusableSlice) Slice() []byte {
	return s.slice
}

// Shrink or grow the slice; length cannot exceed Cap()
func (s *ReusableSlice) Resize(length int) {
	s.slice = s.slice[0:length]
}

func (s *ReusableSlice) Cap() int {
	return cap(s.slice)
}

type SlicePool struct {
	pool *sync.Pool
}

func NewSlicePool(maxLength int) *SlicePool {
	pool := new(sync.Pool)
	pool.New = func() interface{} {
		s := new(ReusableSlice)
		s.pool = pool
		s.slice = make([]byte, maxLength)
		return s
	}
	return &SlicePool{
		pool: pool,
	}
}

func (p *SlicePool) Get() *ReusableSlice {
	s := p.pool.Get().(*ReusableSlice)
	s.Resize(s.Cap())
	s.counter = 1
	return s
}
