package history

import (
	"fmt"
	"sync"
)

type QueueHistory struct {
	Size int
	m    []HistoryItem
	mux  sync.Mutex
}

func MakeQueueHistory(s int) *QueueHistory {
	return &QueueHistory{Size: s, m: make([]HistoryItem, 0, s)}
}

func (l *QueueHistory) Add(val HistoryItem) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if len(l.m) >= l.Size {
		l.m = l.m[1:]
	}

	l.m = append(l.m, val)
}

func (l *QueueHistory) Len() int {
	l.mux.Lock()
	defer l.mux.Unlock()
	return len(l.m)
}

func (l *QueueHistory) Ends() (HistoryItem, HistoryItem, error) {
	old, err := l.Oldest()
	if err != nil {
		return nil, nil, err
	}

	new, err := l.Latest()
	if err != nil {
		return nil, nil, err
	}

	return old, new, nil
}

func (l *QueueHistory) Latest() (HistoryItem, error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if len(l.m) > 0 {
		return l.m[len(l.m)-1], nil
	}
	return nil, fmt.Errorf("empty log")
}

func (l *QueueHistory) Oldest() (HistoryItem, error) {
	l.mux.Lock()
	defer l.mux.Unlock()
	if len(l.m) > 0 {
		return l.m[0], nil
	}
	return nil, fmt.Errorf("empty log")
}
