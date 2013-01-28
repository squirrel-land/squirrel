packet simpleModel

import (
    "time"
    "sync/atomic"
)

// Thread-safe leaky bucket
type leakyBucket struct {
    BucketSize int32
    OutResolution time.Duration
    OutPerSecond int32

    bucket int32
    started bool
}

func (this *leakyBucket) In(size int32) bool {
    if !this.started {
        return false
    }
    if atomic.LoadInt32(&this.bucket) > this.BucketSize {
        return false
    }
    atomic.AddInt32(&this.bucket, size)
    return true
}

func (this *leakyBucket) Go() {
    this.bucket = 0
    this.started = true
    go func() {
        for {
            atomic.AddInt32(&this.bucket, -int32(this.OutPerSecond * this.OutResolution.Seconds()))
            time.Sleep(this.OutResolution)
        }
    }
}
