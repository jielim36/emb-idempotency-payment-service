package redis

import (
	"sync"
	"time"
)

// Simulate redis
type LockManager struct {
	locks sync.Map // key: idempotencyKey(transactionID), value: *sync.Mutex
}

func NewLockManager() *LockManager {
	return &LockManager{}
}

func (lm *LockManager) TryLock(key string) (*sync.Mutex, bool) {
	mu := &sync.Mutex{}
	actual, _ := lm.locks.LoadOrStore(key, mu)
	lock := actual.(*sync.Mutex)

	locked := make(chan struct{}, 1)

	go func() {
		lock.Lock()
		locked <- struct{}{}
	}()

	// try to get lock，100ms timeout
	select {
	case <-locked:
		return lock, true
	case <-time.After(100 * time.Millisecond):
		return nil, false
	}
}

func (lm *LockManager) Unlock(key string) {
	if v, ok := lm.locks.Load(key); ok {
		v.(*sync.Mutex).Unlock()
	}
}
