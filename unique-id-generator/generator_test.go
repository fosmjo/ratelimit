package generator_test

import (
	"fmt"
	"testing"
	"time"

	generator "github.com/fosmjo/system-design-interview/unique-id-generator"
)

func TestGenerator_GenerateID(t *testing.T) {
	g := newGenerator()
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
	g := newGenerator()
	id, err := g.GenerateID()
	if err != nil {
		t.Errorf("GenerateID error: %v", err)
	}

	timeOfID := g.TimeOfID(id)
	if !timeOfID.Equal(_time) {
		t.Errorf("TimeOfID error")
	}
}

func newGenerator() *generator.Generator {
	config, err := generator.NewConfig()
	if err != nil {
		err = fmt.Errorf("NewConfig error: %w", err)
		panic(err)
	}

	g, err := config.NewGenerator(&clock{}, 1, 1)
	if err != nil {
		err = fmt.Errorf("NewGenerator error: %w", err)
		panic(err)
	}

	return g
}

type clock struct{}

var _time = time.Date(2022, time.June, 19, 0, 0, 0, 0, time.UTC)

func (c *clock) Now() time.Time {
	return _time
}
