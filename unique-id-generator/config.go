package generator

import (
	"fmt"
	"time"
)

type Config struct {
	epoch int64 // unix timestamp in milliseconds

	timestampBits    int
	dataCenterIDBits int
	machineIDBits    int
	sequenceBits     int

	timestampShiftBits int

	maxTimestamp    int64
	maxDataCenterID int64
	maxMachineID    int64
	maxSequence     int64
}

func (c *Config) validate() error {
	if c.timestampBits < 0 {
		return fmt.Errorf("timestampBits must be greater than or equal to 0")
	}

	if c.dataCenterIDBits < 0 {
		return fmt.Errorf("dataCenterIDBits must be greater than or equal to 0")
	}

	if c.machineIDBits < 0 {
		return fmt.Errorf("machineIDBits must be greater than or equal to 0")
	}

	if c.sequenceBits < 0 {
		return fmt.Errorf("sequenceBits must be greater than or equal to 0")
	}

	totalBits := c.timestampBits + c.dataCenterIDBits + c.machineIDBits + c.sequenceBits
	if totalBits != 63 {
		return fmt.Errorf("total bits must be  63")
	}

	if c.epoch < 0 {
		return fmt.Errorf("epoch must be >= 0")
	}

	return nil
}

func defaultConfig() *Config {
	return &Config{
		timestampBits:    41,
		dataCenterIDBits: 5,
		machineIDBits:    5,
		sequenceBits:     12,
		epoch:            0,
	}
}

type Option interface {
	apply(c *Config)
}

type optionFunc func(c *Config)

func (f optionFunc) apply(c *Config) {
	f(c)
}

func EpochTimeOption(t time.Time) Option {
	return optionFunc(func(c *Config) {
		c.epoch = t.UnixMilli()
	})
}

func TimestampBitsOption(bits int) Option {
	return optionFunc(func(c *Config) {
		c.timestampBits = bits
	})
}

func DataCenterIDBitsOption(bits int) Option {
	return optionFunc(func(c *Config) {
		c.dataCenterIDBits = bits
	})
}

func MachineIDBitsOption(bits int) Option {
	return optionFunc(func(c *Config) {
		c.machineIDBits = bits
	})
}

func SequenceBitsOption(bits int) Option {
	return optionFunc(func(c *Config) {
		c.sequenceBits = bits
	})
}

func NewConfig(opts ...Option) (*Config, error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt.apply(config)
	}
	if err := config.validate(); err != nil {
		return nil, err
	}

	config.timestampShiftBits = config.dataCenterIDBits + config.machineIDBits + config.sequenceBits
	config.maxTimestamp = -1 ^ (-1 << config.timestampBits)
	config.maxDataCenterID = -1 ^ (-1 << config.dataCenterIDBits)
	config.maxMachineID = -1 ^ (-1 << config.machineIDBits)
	config.maxSequence = -1 ^ (-1 << config.sequenceBits)

	return config, nil
}

func (c *Config) NewGenerator(
	clock Clock, dataCenterID, machineID int64,
) (g *Generator, err error) {
	if dataCenterID > c.maxDataCenterID {
		return nil, ErrDataCenterIDOverflow
	}
	if machineID > c.maxMachineID {
		return nil, ErrMachineIDOverflow
	}

	shiftedDataCenterID := dataCenterID << (c.machineIDBits + c.sequenceBits)
	shiftedMachineID := machineID << c.sequenceBits

	g = &Generator{
		clock:               clock,
		config:              c,
		dataCenterID:        dataCenterID,
		machineID:           machineID,
		shiftedDataCenterID: shiftedDataCenterID,
		shiftedMachineID:    shiftedMachineID,
		lastSeqResetTs:      -1,
	}

	return
}

func (c *Config) realTimestampOfID(id int64) int64 {
	return (id >> c.timestampShiftBits) + c.epoch
}

func (c *Config) shiftedTimestamp(realTs int64) (int64, error) {
	realTs = realTs - c.epoch
	if realTs < 0 || realTs > c.maxTimestamp {
		return 0, ErrTimestampOverflow
	}

	shiftedTs := realTs << c.timestampShiftBits

	return shiftedTs, nil
}
