package globals

import (
	"math"
	"sync"
)

type IdGenertor interface {
	GetUint32() uint32
}

type cyclicIdGenertor struct {
	sn    uint32
	ended bool
	mutex sync.Mutex
}

func NewIdGenertor() IdGenertor {
	return &cyclicIdGenertor{}
}

func (gen *cyclicIdGenertor) GetUint32() uint32 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	if gen.ended {
		defer func() {
			gen.ended = false
		}()
		gen.sn = 0
		return gen.sn
	}

	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	} else {
		gen.ended = true
	}

	return id
}
