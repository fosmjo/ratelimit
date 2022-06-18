package ratelimit

import "time"

type Colck interface {
	Now() time.Time
}
