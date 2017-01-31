package history

import (
	"fmt"
	"sync"
	"time"
)

// store and query items by time
// each item in UniqueHistory is unique
type UniqueHistory struct {
	length time.Duration             // constraint on newest time - oldest time
	m      map[HistoryItem]time.Time // item -> time
	t      map[time.Time]HistoryItem // time -> item
	times  []time.Time               // sorted slice of keys in map
	mux    sync.Mutex                // for thread-safeness
}

func MakeUniqueHistory(d time.Duration) *UniqueHistory {
	return &UniqueHistory{
		length: d,
		m:      make(map[HistoryItem]time.Time),
		t:      make(map[time.Time]HistoryItem),
		times:  make([]time.Time, 0),
	}
}

func (l *UniqueHistory) UpdateDuration(d time.Duration) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.length = d
}

func (l *UniqueHistory) Len() int {
	return len(l.times)
}

func (l *UniqueHistory) Add(t time.Time, it HistoryItem) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if t, ok := l.m[it]; ok {
		// duplicate value. remove old one.
		delete(l.m, it)
		delete(l.t, t)
		// remove from times
		// TODO binary search
		for i, v := range l.times {
			if v == t {
				l.times = append(l.times[:i], l.times[i+1:]...)
				break
			}
		}
	}

	l.times = append(l.times, t)
	l.m[it] = t
	l.t[t] = it

	if len(l.times) != len(l.m) {
		err := fmt.Errorf("UniqueHistory in inconsistent state: %v %v", len(l.times), len(l.m))
		panic(err)
	}

	lastTime := l.times[len(l.times)-1]
	// remove older, keep at least 100
	for len(l.times) > 100 && lastTime.Sub(l.times[0]) > l.length {
		rem := l.times[0]
		seq, _ := l.t[rem] // seq
		delete(l.t, rem)
		delete(l.m, seq)
		l.times = l.times[1:]
	}
}

// last item before given time and time it was logged
func (l *UniqueHistory) Before(wanted time.Time) (HistoryItem, time.Time, error) {
	var then time.Time

	if len(l.times) == 0 {
		return nil, time.Now(), fmt.Errorf("empty log")
	}

	if wanted.Before(l.times[0]) {
		return l.t[l.times[0]], l.times[0], fmt.Errorf("wanted time before log start")
	}

	for _, t := range l.times {
		if t.After(wanted) {
			return l.t[then], then, nil
		} else {
			then = t
		}
	}

	lastTime := l.times[len(l.times)-1]
	return l.t[lastTime], lastTime, nil
}

func (l *UniqueHistory) NumItemsBetween(start time.Time, end time.Time) (int, error) {
	count := 0
	for _, t := range l.times {
		if !t.Before(end) {
			return count, nil
		} else if !t.Before(start) {
			count++
		}
	}

	return count, nil
}