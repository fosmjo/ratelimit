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

	sequence         int64
	lastSeqResetTime time.Time
	seqMu            sync.Mutex
}

func (g *Generator) GenerateID() (id int64, err error) {
	t, seq := g.nextSeqID()
	shiftedTs, err := g.config.shiftedTimestamp(t.UnixMilli())
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

func (g *Generator) SeqOfID(id int64) int64 {
	return id & g.config.maxSequence
}

func (g *Generator) nextSeqID() (time.Time, int64) {
	g.seqMu.Lock()
	defer g.seqMu.Unlock()

	for {
		now := g.clock.Now().UTC()

		if now.UnixMilli() > g.lastSeqResetTime.UnixMilli() {
			g.sequence = 0
			g.lastSeqResetTime = now
		} else if g.sequence >= g.config.maxSequence {
			time.Sleep(now.Sub(g.lastSeqResetTime))
			continue
		} else {
			g.sequence++
		}

		return now, g.sequence
	}
}
