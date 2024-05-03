package loaddata

import (
	"errors"
	"reflect"
	"testing"
)

type MockDataLoader struct {
	cache map[int]string
	db    map[int]string
}

func (m *MockDataLoader) GetFromCache(key int) (string, bool) {
	value, found := m.cache[key]
	if key == 5 { // Simulate a cache miss
		return "", false
	}
	return value, found
}

func (m *MockDataLoader) LoadFromDB(key int) error {
	value, found := m.db[key]
	if !found {
		return errors.New("key not found in DB")
	}
	m.cache[key] = value
	return nil
}

type MockLockManager struct{}

func (m *MockLockManager) TryLock(key int) func() {
	return func() {}
}

func TestGetOrLoad(t *testing.T) {
	type args[K comparable, V any] struct {
		dataLoader  DataLoader[int, string]
		lockManager LockManager[int]
		key         K
	}
	type testCase[K comparable, V any] struct {
		name    string
		args    args[K, V]
		want    V
		wantErr bool
	}
	tests := []testCase[int, string]{
		{
			name: "Data found in cache",
			args: args[int, string]{
				dataLoader: &MockDataLoader{
					cache: map[int]string{1: "value1"},
					db:    map[int]string{},
				},
				lockManager: &MockLockManager{},
				key:         1,
			},
			want:    "value1",
			wantErr: false,
		},
		{
			name: "Data not found in cache",
			args: args[int, string]{
				dataLoader: &MockDataLoader{
					cache: map[int]string{},
					db:    map[int]string{2: "value2"},
				},
				lockManager: &MockLockManager{},
				key:         2,
			},
			want:    "value2",
			wantErr: false,
		},
		{
			name: "Data not found in cache and DB",
			args: args[int, string]{
				dataLoader: &MockDataLoader{
					cache: map[int]string{},
					db:    map[int]string{},
				},
				lockManager: &MockLockManager{},
				key:         3,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Data not found in cache, load from DB success but cache miss",
			args: args[int, string]{
				dataLoader: &MockDataLoader{
					cache: map[int]string{},
					db:    map[int]string{5: "value5"},
				},
				lockManager: &MockLockManager{},
				key:         5,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOrLoad(tt.args.key, tt.args.dataLoader, tt.args.lockManager)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrLoad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrLoad() got = %v, want %v", got, tt.want)
			}
		})
	}
}
