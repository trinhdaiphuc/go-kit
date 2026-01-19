package sonyflake

import (
	"time"

	"github.com/sony/sonyflake/v2"
)

type Option func(*sonyflake.Settings)

func WithStartTime(startTime time.Time) Option {
	return func(settings *sonyflake.Settings) {
		settings.StartTime = startTime
	}
}

func WithTimeUnit(timeUnit time.Duration) Option {
	return func(settings *sonyflake.Settings) {
		settings.TimeUnit = timeUnit
	}
}

func WithBitsSequence(bitsSequence int) Option {
	return func(settings *sonyflake.Settings) {
		settings.BitsSequence = bitsSequence
	}
}

func WithBitsMachineID(bitsMachineID int) Option {
	return func(settings *sonyflake.Settings) {
		settings.BitsMachineID = bitsMachineID
	}
}

func WithMachineID(machineID func() (int, error)) Option {
	return func(settings *sonyflake.Settings) {
		settings.MachineID = machineID
	}
}

func WithCheckMachineID(checkMachineID func(int) bool) Option {
	return func(settings *sonyflake.Settings) {
		settings.CheckMachineID = checkMachineID
	}
}
