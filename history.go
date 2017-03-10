package history

import (
	"fmt"
	"sync"
	"time"
)

type History struct {
	length time.Duration             // constraint on newest time - oldest time
	t      map[time.Time]HistoryItem // time -> item
	times  []time.Time               // sorted slice of keys in map
	mux    sync.Mutex                // for thread-safeness
}

func MakeHistory(d time.Duration) *History {
	return &History{
		length: d,
		t:      make(map[time.Time]HistoryItem),
		times:  make([]time.Time, 0),
	}
}

func (l *History) UpdateDuration(d time.Duration) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.length = d
}

func (l *History) Len() int {
	l.mux.Lock()
	defer l.mux.Unlock()

	return len(l.times)
}

func (l *History) Add(t time.Time, it HistoryItem) {
	l.mux.Lock()
	defer l.mux.Unlock()

	l.times = append(l.times, t)
	l.t[t] = it

	if len(l.times) != len(l.t) {
		err := fmt.Errorf("History in inconsistent state: %v %v", len(l.times), len(l.t))
		panic(err)
	}

	lastTime := l.times[len(l.times)-1]
	// remove older, keep at least 100
	for len(l.times) > 100 && lastTime.Sub(l.times[0]) > l.length {
		rem := l.times[0]
		delete(l.t, rem)
		l.times = l.times[1:]
	}
}

// last item before given time and time it was logged
func (l *History) Before(wanted time.Time) (HistoryItem, time.Time, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if len(l.times) == 0 {
		return nil, time.Now(), fmt.Errorf("empty log")
	}

	if wanted.Before(l.times[0]) {
		return l.t[l.times[0]], l.times[0], fmt.Errorf("wanted time before log start")
	}

	then := binsearch(wanted, l.times)
	return l.t[then], then, nil
}

func binsearchindex(wanted time.Time, times []time.Time, index int) int {
	if len(times) <= 8 {
		return linsearchindex(wanted, times, index)
	}

	mid := len(times) / 2

	if diff := wanted.Sub(times[mid]); diff > 0 {
		return binsearchindex(wanted, times[mid:], index+mid)
	} else if diff < 0 {
		return binsearchindex(wanted, times[:mid], index)
	} else {
		return index+mid
	}
}

func linsearchindex(wanted time.Time, times []time.Time, index int) int {
	count := 0
	for _, t := range times {
		if t.After(wanted) {
			return index+count
		} else {
			count += 1
		}
	}
	if count>=len(times) {
		return index+len(times)-1
	} else {
		return index+count
	}
}

func binsearch(wanted time.Time, times []time.Time) time.Time {
	if len(times) <= 8 {
		return linsearch(wanted, times)
	}

	mid := len(times) / 2

	if diff := wanted.Sub(times[mid]); diff > 0 {
		return binsearch(wanted, times[mid:])
	} else if diff < 0 {
		return binsearch(wanted, times[:mid])
	} else {
		return times[mid]
	}
}

func linsearch(wanted time.Time, times []time.Time) time.Time {
	then := times[0]
	for _, t := range times {
		if t.After(wanted) {
			return then
		} else {
			then = t
		}
	}
	return then
}

func (l *History) NumItemsBetween(start time.Time, end time.Time) (int, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if len(l.times) == 0 {
		return 0, fmt.Errorf("empty log")
	}

	startIndex := binsearchindex(start, l.times, 0)
	endIndex := binsearchindex(end, l.times, 0)
	count := 0
	if l.times[startIndex] == start {
		count = endIndex-startIndex-1
	} else {
		count = endIndex - startIndex
	}
	if count < 0 {
		count = 0
	}
	return count, nil
}

func (l *History) ItemsBetween(start time.Time, end time.Time) ([]HistoryItemWithTime, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	its := make([]HistoryItemWithTime, 0)

	startIndex := binsearchindex(start, l.times, 0)
	endIndex := binsearchindex(end, l.times, 0)

	if l.times[startIndex] == start {
		startIndex += 1
	}
	
	for i:= startIndex; i<=endIndex; i++ {
		its = append(its, HistoryItemWithTime{Time: l.times[i], Item: l.t[l.times[i]]})
	}

	return its, nil
}

func (l *History) AvgBetween(
	from time.Time,
	to time.Time,
	zero HistoryItem,
	sum func(a HistoryItem, b HistoryItem) HistoryItem,
	div func(a HistoryItem, n int) HistoryItem,
) (HistoryItem, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	cum := zero
	count := 0
	// TODO binary search
	for _, t := range l.times {
		if t.After(from) {
			if t.Before(to) {
				cum = sum(cum, l.t[t])
				count++
			} else {
				break
			}
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("timed log: no values to avg")
	}

	return div(cum, count), nil
}
