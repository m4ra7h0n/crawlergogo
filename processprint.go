package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type ProcessBar struct {
	lock    sync.Mutex
	start   int64
	data    int64
	counter int64
	total   int64
	success int32
	fail    int32
}

func NewProcessBar(total int64) *ProcessBar {
	return &ProcessBar{
		start:   time.Now().UnixNano(),
		total:   total,
		data:    0,
		counter: 0,
		success: 0,
		fail:    0,
	}
}

func (pb *ProcessBar) SetTotal(total int64) {
	pb.lock.Lock()
	defer pb.lock.Unlock()
	pb.total = total
}

func (pb *ProcessBar) PrintAccumulator(data int64, success int32, fail int32, header string) {
	pb.lock.Lock()
	defer pb.lock.Unlock()

	pb.data += data
	pb.counter++
	pb.success += success
	pb.fail += fail

	barN := int((float64(pb.data) / float64(pb.total) * 100) / 2)
	dur := time.Since(time.Unix(0, pb.start)).Seconds()
	remainingDur := float64(pb.total-pb.data) * (dur / float64(pb.counter))

	h, remainder := divmod(int(dur), 60*60)
	m, s := divmod(remainder, 60)
	l_h, l_remainder := divmod(int(remainingDur), 60*60)
	l_m, l_s := divmod(l_remainder, 60)

	fmt.Printf("\r%s: %3d%% (%d/%d) : %s%s  %02d:%02d:%02d (预计剩余:%02d:%02d:%02d) (success:%d, fail:%d)",
		header,
		int((pb.data*100)/pb.total),
		pb.data, pb.total,
		strings.Repeat("▋", barN),
		strings.Repeat("_", 50-barN),
		h, m, s,
		l_h, l_m, l_s,
		pb.success, pb.fail)
	time.Sleep(50 * time.Millisecond)
}

func divmod(numerator int, denominator int) (int, int) {
	return numerator / denominator, numerator % denominator
}
