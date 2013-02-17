package models

import (
	"sync/atomic"
	"time"
)

// Thread-safe leaky bucket
type leakyBucket struct {
	BucketSize        int32
	OutResolution     int32 // in milliseconds
	OutPerMilliSecond int32

	bucket  int32
	started bool
}

// by far (before Go 1.1), int is 32 bit no matter what arch it is. After Go 1.1 release, int64 should be used here, since int is going to be 64 bit on 64 bit arch and 64 bit machine would be the most common platform that Master runs on.

func (this *leakyBucket) In(size int) bool {
	if atomic.LoadInt32(&this.bucket) > this.BucketSize {
		return false
	}
	atomic.AddInt32(&this.bucket, int32(size))
	return true
}

func (this *leakyBucket) Go() {
	go func() {
		sleepTime := time.Duration(this.OutResolution) * time.Millisecond
		for {
			if atomic.LoadInt32(&this.bucket) > 0 {
				atomic.AddInt32(&this.bucket, -int32(atomic.LoadInt32(&this.OutPerMilliSecond)*this.OutResolution))
			}
			time.Sleep(sleepTime)
		}
	}()
}

func (this *leakyBucket) UpdateOutRate(outPerMilliSecond int) {
	atomic.StoreInt32(&this.OutPerMilliSecond, int32(outPerMilliSecond))
}
