package generator

import (
	"fmt"
	"sync"
	"time"
)

var (
	ErrTimestampOverflow    = fmt.Errorf("timestamp overflow")
	ErrDataCenterIDOverflow = fmt.Errorf("data center id overflow")
	ErrMachineIDOverflow    = fmt.Errorf("machine id overflow")
	ErrSequenceOverflow     = fmt.Errorf("sequence overflow")
)

type Clock interface {
	Now() time.Time
}

type Generator struct {
	clock  Clock
	config *Config

	dataCenterID        int64
	machineID           int64
	shiftedDataCenterID int64
	shiftedMachineID    int64

	sequence       int64
	lastSeqResetTs int64
	seqMu          sync.Mutex
}

func (g *Generator) GenerateID() (id int64, err error) {
	shiftedTs, err := g.config.shiftedTimestamp(g.clock.Now().UnixMilli())
	if err != nil {
		return 0, err
	}

	seq, err := g.nextSequence(shiftedTs)
	if err != nil {
		return 0, err
	}

	id = shiftedTs | g.shiftedDataCenterID | g.shiftedMachineID | seq

	return
}

func (g *Generator) TimeOfID(id int64) time.Time {
	ts := g.config.realTimestampOfID(id)
	return time.Unix(int64(ts/1000), int64(ts%1000)*1_000_000)
}

func (g *Generator) nextSequence(ts int64) (int64, error) {
	g.seqMu.Lock()
	defer g.seqMu.Unlock()

	if ts > g.lastSeqResetTs {
		g.sequence = 0
		g.lastSeqResetTs = ts
	}

	if g.sequence >= g.config.maxSequence {
		return 0, ErrSequenceOverflow
	}

	g.sequence++

	return g.sequence, nil
}
