package loaddata

import (
	"fmt"
	"sync"
)

// DataLoader interface for loading data from database or any source.
type DataLoader[K comparable, V any] interface {
	GetFromCache(key K) (V, bool)
	LoadFromDB(key K) error
}

// LockManager represents an abstraction over locking mechanisms.
type LockManager[K comparable] interface {
	TryLock(key K) func() // Returns a function to unlock
}

// GetOrLoad attempts to retrieve a value with the specified key from the cache.
// If the value is not found in the cache, it is loaded from the database.
func GetOrLoad[K comparable, V any](key K, dataLoader DataLoader[K, V], lockManager LockManager[K]) (V, error) {
	result, found := dataLoader.GetFromCache(key)
	if found {
		return result, nil
	}

	fmt.Println("Data not found in dataLoader, key:", key)

	// Try to acquire the lock
	unlock := lockManager.TryLock(key)
	defer unlock() // Ensure the lock is always released

	// The lock might have been acquired after another routine loaded the data,
	// so check the dataLoader again to avoid unnecessary loading.
	result, found = dataLoader.GetFromCache(key)
	if !found {
		// Value still not in cache, load it from the database
		if err := dataLoader.LoadFromDB(key); err != nil {
			return result, fmt.Errorf("error on reloading data for key %v: %v", key, err)
		}
		result, found = dataLoader.GetFromCache(key)
		if !found {
			// Failsafe: in case loading to cached was unsuccessful
			return result, fmt.Errorf("failed to load data for key %v", key)
		}
	}

	return result, nil
}

var (
	_ LockManager[any] = &SimpleLockManager{}
)

// SimpleLockManager implementation of the lock manager using a simple mutex.
type SimpleLockManager struct {
	mutex sync.Mutex
}

func (l *SimpleLockManager) TryLock(key any) func() {
	l.mutex.Lock()
	return func() {
		l.mutex.Unlock()
	}
}
