package sonyflake

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sony/sonyflake/v2"
	"github.com/trinhdaiphuc/go-kit/network"
)

var (
	once     = sync.Once{}
	instance *sonyflake.Sonyflake
)

func NewSonyflake(opts ...Option) (*sonyflake.Sonyflake, error) {
	settings := sonyflake.Settings{}
	for _, opt := range opts {
		opt(&settings)
	}

	if settings.MachineID == nil {
		settings.MachineID = defaultMachineID
	}
	var err error
	once.Do(func() {
		instance, err = sonyflake.New(settings)
	})

	return instance, err
}

// defaultMachineID returns the last two octets of the local IP address as a 16-bit machine ID.
// Example: IP 192.168.1.23 -> machineID = (1 << 8) | 23
func defaultMachineID() (int, error) {
	ip, err := network.LocalIP()
	if err != nil {
		return 0, fmt.Errorf("Get local ip failed: %w", err)
	}

	if len(ip) < 4 {
		return 0, errors.New("Invalid local ip")
	}
	machineID := (int(ip[2]) << 8) | int(ip[3])
	return machineID, nil
}

func NextID() (int64, error) {
	return instance.NextID()
}

func NextIDString() (string, error) {
	id, err := instance.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func ToTime(id int64) time.Time {
	return instance.ToTime(id)
}

func Decompose(id int64) map[string]int64 {
	return instance.Decompose(id)
}

func Compose(t time.Time, sequence, machineID int) (int64, error) {
	return instance.Compose(t, sequence, machineID)
}
