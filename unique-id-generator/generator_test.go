package generator_test

import (
	"fmt"
	"testing"
	"time"

	generator "github.com/fosmjo/system-design-interview/unique-id-generator"
)

func TestGenerator_GenerateID(t *testing.T) {
	g := newGenerator(&realClock{})
	id, err := g.GenerateID()
	if err != nil {
		t.Errorf("GenerateID error: %v", err)
	}

	id2, err := g.GenerateID()
	if err != nil {
		t.Errorf("GenerateID error 2: %v", err)
	}

	if id2 <= id {
		t.Errorf("GenerateID error: id2 <= id")
	}
}

func TestGenerator_TimeOfID(t *testing.T) {
	g := newGenerator(&dummyClock{})
	id, err := g.GenerateID()
	if err != nil {
		t.Errorf("GenerateID error: %v", err)
	}

	timeOfID := g.TimeOfID(id)
	if !timeOfID.Equal(_time) {
		t.Errorf("TimeOfID error")
	}
}

func BenchmarkGenerateID(b *testing.B) {
	g := newGenerator(&realClock{})

	for i := 0; i < b.N; i++ {
		g.GenerateID()
	}
}

func BenchmarkGenerateIDParallel(b *testing.B) {
	g := newGenerator(&realClock{})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.GenerateID()
		}
	})
}

func newGenerator(clock generator.Clock) *generator.Generator {
	config, err := generator.NewConfig()
	if err != nil {
		err = fmt.Errorf("NewConfig error: %w", err)
		panic(err)
	}

	g, err := config.NewGenerator(clock, 1, 1)
	if err != nil {
		err = fmt.Errorf("NewGenerator error: %w", err)
		panic(err)
	}

	return g
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now()
}

var _time = time.Date(2022, time.June, 19, 0, 0, 0, 0, time.UTC)

type dummyClock struct{}

func (c *dummyClock) Now() time.Time {
	return _time
}
