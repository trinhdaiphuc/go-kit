package cacheredis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type Data struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func newRedisClientMock[K comparable, V any](loader cache.Loader[K, V]) (cache.Store[K, V], redismock.ClientMock) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisCache[K, V](client, WithLoader[K, V](loader), WithPrefix[K, V]("test"))
	return repo, mock
}

type loaderSuccess struct{}

func (l *loaderSuccess) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	return &Data{Name: "John Doe", Value: 100}, nil
}

func (l *loaderSuccess) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	return map[string]*Data{
		"field1": {Name: "John Doe", Value: 100},
		"field2": {Name: "Jane Doe", Value: 200},
	}, nil
}

type loaderFailed struct{}

func (l *loaderFailed) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	return nil, errors.New("failed")
}

func (l *loaderFailed) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	return nil, errors.New("failed")
}

func Test_redisCache_Get(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx    context.Context
		key    K
		loader cache.Loader[K, V]
	}
	type testCase[K comparable, V any] struct {
		name    K
		args    args[K, V]
		mock    func(mock redismock.ClientMock)
		want    V
		wantErr assert.ErrorAssertionFunc
	}
	var (
		valueStr    = "{\"name\":\"John Doe\",\"value\":100}"
		value       = &Data{Name: "John Doe", Value: 100}
		errInternal = errors.New("internal")
	)
	tests := []testCase[string, *Data]{
		{
			name: "Get value successfully",
			args: args[string, *Data]{
				ctx: context.Background(),
				key: "key",
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("test:key").SetVal(valueStr)
			},
			want:    value,
			wantErr: assert.NoError,
		},
		{
			name: "Get value not found, load successfully",
			args: args[string, *Data]{
				ctx:    context.Background(),
				key:    "key",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("test:key").RedisNil()
			},
			want:    value,
			wantErr: assert.NoError,
		},
		{
			name: "Get value not found, load failed",
			args: args[string, *Data]{
				ctx:    context.Background(),
				key:    "key",
				loader: &loaderFailed{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("test:key").RedisNil()
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Get value not found, loader not found",
			args: args[string, *Data]{
				ctx:    context.Background(),
				key:    "key",
				loader: nil,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("test:key").RedisNil()
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Get value internal error",
			args: args[string, *Data]{
				ctx: context.Background(),
				key: "key",
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectGet("test:key").SetErr(errInternal)
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock(tt.args.loader)
			tt.mock(repoMock)
			got, err := repo.Get(tt.args.ctx, tt.args.key)
			if !tt.wantErr(t, err, "Get(%v, %v)", tt.args.ctx, tt.args.key) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.key)
		})
	}
}

func Test_redisCache_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx  context.Context
		key  K
		data V
	}
	type testCase[K comparable, V any] struct {
		name    K
		args    args[K, V]
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}
	var (
		value = &Data{Name: "John Doe", Value: 100}
	)
	tests := []testCase[string, *Data]{
		{
			name: "Set value successfully",
			args: args[string, *Data]{
				ctx:  context.Background(),
				key:  "key",
				data: value,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSet("test:key", "{\"name\":\"John Doe\",\"value\":100}", 0).SetVal("1")
			},
			wantErr: assert.NoError,
		},
		{
			name: "Set value internal error",
			args: args[string, *Data]{
				ctx:  context.Background(),
				key:  "key",
				data: value,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSet("test:key", "{\"name\":\"John Doe\",\"value\":100}", 0).SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
		{
			name: "Marshal error",
			args: args[string, *Data]{
				ctx:  context.Background(),
				key:  "key",
				data: nil,
			},
			mock: func(mock redismock.ClientMock) {
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.Set(tt.args.ctx, tt.args.key, tt.args.data, 0)
			tt.wantErr(t, err, "Set(%v, %v, %v, 0)", tt.args.ctx, tt.args.key, tt.args.data)
		})
	}
}

func Test_redisCache_SetNX(t *testing.T) {
	type args struct {
		ctx  context.Context
		key  string
		data *Data
	}
	var (
		value = &Data{Name: "John Doe", Value: 100}
	)
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "SetNX value successfully",
			args: args{
				ctx:  context.Background(),
				key:  "key",
				data: value,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSetNX("test:key", "{\"name\":\"John Doe\",\"value\":100}", 0).SetVal(true)
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "SetNX value failed",
			args: args{
				ctx:  context.Background(),
				key:  "key",
				data: value,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSetNX("test:key", "{\"name\":\"John Doe\",\"value\":100}", 0).SetVal(false)
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "SetNX value internal error",
			args: args{
				ctx:  context.Background(),
				key:  "key",
				data: value,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectSetNX("test:key", "{\"name\":\"John Doe\",\"value\":100}", 0).SetErr(errors.New("internal"))
			},
			want:    false,
			wantErr: assert.Error,
		},
		{
			name: "Marshal error",
			args: args{
				ctx:  context.Background(),
				key:  "key",
				data: nil,
			},
			mock: func(mock redismock.ClientMock) {
			},
			want:    false,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			got, err := repo.SetNX(tt.args.ctx, tt.args.key, tt.args.data, 0)
			if !tt.wantErr(t, err, "SetNX(%v, %v, %v, 0)", tt.args.ctx, tt.args.key, tt.args.data) {
				return
			}
			assert.Equalf(t, tt.want, got, "SetNX(%v, %v, %v, 0)", tt.args.ctx, tt.args.key, tt.args.data)
		})
	}
}

func Test_redisCache_Delete(t *testing.T) {
	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Delete value successfully",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "key2"},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectDel("test:key1", "test:key2").SetVal(2)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Delete value internal error",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "key2"},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectDel("test:key1", "test:key2").SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.Delete(tt.args.ctx, tt.args.keys...)
			tt.wantErr(t, err, "Delete(%v, %v)", tt.args.ctx, tt.args.keys)
		})
	}
}

func Test_redisCache_Incr(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value int64
	}
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		want    int64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Incr value successfully",
			args: args{
				ctx:   context.Background(),
				key:   "key",
				value: 10,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectIncrBy("test:key", 10).SetVal(20)
			},
			want:    20,
			wantErr: assert.NoError,
		},
		{
			name: "Incr value internal error",
			args: args{
				ctx:   context.Background(),
				key:   "key",
				value: 10,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectIncrBy("test:key", 10).SetErr(errors.New("internal"))
			},
			want:    0,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			got, err := repo.Incr(tt.args.ctx, tt.args.key, tt.args.value)
			if !tt.wantErr(t, err, "Incr(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.value) {
				return
			}
			assert.Equalf(t, tt.want, got, "Incr(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.value)
		})
	}
}

func Test_redisCache_Expire(t *testing.T) {
	type args struct {
		ctx        context.Context
		key        string
		expireTime time.Duration
	}
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Expire value successfully",
			args: args{
				ctx:        context.Background(),
				key:        "key",
				expireTime: 10 * time.Second,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectExpire("test:key", 10*time.Second).SetVal(true)
			},
			wantErr: assert.NoError,
		},
		{
			name: "Expire value internal error",
			args: args{
				ctx:        context.Background(),
				key:        "key",
				expireTime: 10 * time.Second,
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectExpire("test:key", 10*time.Second).SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.Expire(tt.args.ctx, tt.args.key, tt.args.expireTime)
			tt.wantErr(t, err, "Expire(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.expireTime)
		})
	}
}

func Test_redisCache_HSet(t *testing.T) {
	type args struct {
		ctx     context.Context
		key     string
		keyVals []cache.KeyVal[string, *Data]
	}
	var (
		value = &Data{Name: "John Doe", Value: 100}
	)
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "HSet value successfully",
			args: args{
				ctx: context.Background(),
				key: "key",
				keyVals: []cache.KeyVal[string, *Data]{
					{Key: "field", Value: value},
				},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHSet("test:key", "field", "{\"name\":\"John Doe\",\"value\":100}").SetVal(1)
			},
			wantErr: assert.NoError,
		},
		{
			name: "HSet value internal error",
			args: args{
				ctx: context.Background(),
				key: "key",
				keyVals: []cache.KeyVal[string, *Data]{
					{Key: "field", Value: value},
				},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHSet("test:key", "field", "{\"name\":\"John Doe\",\"value\":100}").SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
		{
			name: "Marshal error",
			args: args{
				ctx: context.Background(),
				key: "key",
				keyVals: []cache.KeyVal[string, *Data]{
					{Key: "field", Value: nil},
				},
			},
			mock: func(mock redismock.ClientMock) {
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.HSet(tt.args.ctx, tt.args.key, tt.args.keyVals...)
			tt.wantErr(t, err, "HSet(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.keyVals)
		})
	}
}

func Test_redisCache_HGet(t *testing.T) {
	type args struct {
		ctx    context.Context
		key    string
		field  string
		loader cache.Loader[string, *Data]
	}
	var (
		valueStr    = "{\"name\":\"John Doe\",\"value\":100}"
		value       = &Data{Name: "John Doe", Value: 100}
		errInternal = errors.New("internal")
	)
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		want    *Data
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "HGet value successfully",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				field:  "field",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGet("test:key", "field").SetVal(valueStr)
			},
			want:    value,
			wantErr: assert.NoError,
		},
		{
			name: "HGet value not found, load successfully",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				field:  "field",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGet("test:key", "field").RedisNil()
			},
			want:    value,
			wantErr: assert.NoError,
		},
		{
			name: "HGet value not found, load failed",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				field:  "field",
				loader: &loaderFailed{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGet("test:key", "field").RedisNil()
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "HGet value internal error",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				field:  "field",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGet("test:key", "field").SetErr(errInternal)
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](tt.args.loader)
			tt.mock(repoMock)
			got, err := repo.HGet(tt.args.ctx, tt.args.key, tt.args.field)
			if !tt.wantErr(t, err, "HGet(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.field) {
				return
			}
			assert.Equalf(t, tt.want, got, "HGet(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.field)
		})
	}
}

func Test_redisCache_HGetAll(t *testing.T) {
	type args struct {
		ctx    context.Context
		key    string
		loader cache.Loader[string, *Data]
	}
	var (
		valueStr1   = "{\"name\":\"John Doe\",\"value\":100}"
		valueStr2   = "{\"name\":\"Jane Doe\",\"value\":200}"
		value1      = &Data{Name: "John Doe", Value: 100}
		value2      = &Data{Name: "Jane Doe", Value: 200}
		valueMap    = map[string]*Data{"field1": value1, "field2": value2}
		valueMapStr = map[string]string{"field1": valueStr1, "field2": valueStr2}
	)
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		want    map[string]*Data
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "HGetAll value successfully",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:key").SetVal(valueMapStr)
			},
			want:    valueMap,
			wantErr: assert.NoError,
		},
		{
			name: "HGetAll value not found, load successfully",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				loader: &loaderSuccess{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:key").SetVal(map[string]string{})
			},
			want:    valueMap,
			wantErr: assert.NoError,
		},
		{
			name: "HGetAll value not found, load failed",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				loader: &loaderFailed{},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:key").RedisNil()
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "HGetAll value internal error",
			args: args{
				ctx: context.Background(),
				key: "key",
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHGetAll("test:key").SetErr(errors.New("internal"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](tt.args.loader)
			tt.mock(repoMock)
			got, err := repo.HGetAll(tt.args.ctx, tt.args.key)
			if !tt.wantErr(t, err, "HGetAll(%v, %v)", tt.args.ctx, tt.args.key) {
				return
			}
			assert.Equalf(t, tt.want, got, "HGetAll(%v, %v)", tt.args.ctx, tt.args.key)
		})
	}
}

func Test_redisCache_HDel(t *testing.T) {
	type args struct {
		ctx    context.Context
		key    string
		fields []string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "HDel value successfully",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				fields: []string{"field1", "field2"},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHDel("test:key", "field1", "field2").SetVal(2)
			},
			wantErr: assert.NoError,
		},
		{
			name: "HDel value internal error",
			args: args{
				ctx:    context.Background(),
				key:    "key",
				fields: []string{"field1", "field2"},
			},
			mock: func(mock redismock.ClientMock) {
				mock.ExpectHDel("test:key", "field1", "field2").SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.HDel(tt.args.ctx, tt.args.key, tt.args.fields...)
			tt.wantErr(t, err, "HDel(%v, %v, %v)", tt.args.ctx, tt.args.key, tt.args.fields)
		})
	}
}

func Test_redisCache_Ping(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Ping successfully",
			mock: func(mock redismock.ClientMock) {
				mock.ExpectPing().SetVal("PONG")
			},
			wantErr: assert.NoError,
		},
		{
			name: "Ping internal error",
			mock: func(mock redismock.ClientMock) {
				mock.ExpectPing().SetErr(errors.New("internal"))
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			err := repo.Ping(context.Background())
			tt.wantErr(t, err, "Ping()")
		})
	}
}

func Test_redisCache_Close(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(mock redismock.ClientMock)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Close successfully",
			mock: func(mock redismock.ClientMock) {
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, repoMock := newRedisClientMock[string, *Data](nil)
			tt.mock(repoMock)
			repo.Close()
		})
	}
}
