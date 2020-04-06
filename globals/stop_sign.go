package globals

import (
	"fmt"
	"sync"
)

type IStopSign interface {
	Sign() bool
	Signed() bool
	Reset()
	Deal(code string)
	DealCount(code string) uint32
	DealTotal() uint32
	Summary() string
}

var stopSignSummaryTemplate = "signed: %v, " + "map: %v"

type myStopSign struct {
	signed       bool
	dealCountMap map[string]uint32
	rwmutex      sync.RWMutex
}

func NewStopSign() IStopSign {
	ss := &myStopSign{
		dealCountMap: make(map[string]uint32),
	}
	return ss
}

func (ss *myStopSign) Sign() bool {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	if ss.signed {
		return false
	}
	ss.signed = true
	return true
}

func (ss *myStopSign) Signed() bool {
	return ss.signed
}

func (ss *myStopSign) Deal(code string) {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	if !ss.signed {
		return
	}
	if _, ok := ss.dealCountMap[code]; !ok {
		ss.dealCountMap[code] = 1
	} else {
		ss.dealCountMap[code] += 1
	}
}

func (ss *myStopSign) Reset() {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	ss.signed = false
	ss.dealCountMap = make(map[string]uint32)
}

func (ss *myStopSign) DealCount(code string) uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	count, ok := ss.dealCountMap[code]
	if ok {
		return count
	} else {
		return 0
	}
}

func (ss *myStopSign) DealTotal() uint32 {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	var total uint32 = 0
	for _, v := range ss.dealCountMap {
		total += v
	}
	return uint32(total)
}

func (ss *myStopSign) Summary() string {
	ss.rwmutex.Lock()
	defer ss.rwmutex.Unlock()

	mapValues := ""
	for k, v := range ss.dealCountMap {
		mapValues += fmt.Sprintf("%s:%d,", k, v)
	}

	summary := fmt.Sprintf(stopSignSummaryTemplate,
		ss.signed, mapValues)
	return summary
}
