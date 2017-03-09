package history

import (
	"testing"
)

func TestLen(t *testing.T) {
	l := MakeQueueHistory(10)
	l.Add(42)

	if s := l.Len(); s != 1 {
		t.Error("wrong len", s, "expected 1")
	}
}

func TestLenLimit(t *testing.T) {
	l := MakeQueueHistory(10)

	for i := 0; i < 20; i++ {
		l.Add(42)
	}

	if s := l.Len(); s != 10 {
		t.Error("wrong len, got", s, "expected 10")
	}
}

func TestEnds(t *testing.T) {
	l := MakeQueueHistory(10)

	for i := 50; i < 52; i++ {
		l.Add(i)
	}

	old, err := l.Oldest()
	if err != nil {
		t.Errorf("got error on Ends(): %s", err)
	}

	new, err := l.Latest()
	if err != nil {
		t.Errorf("got error on Ends(): %s", err)
	}

	oldest := old.(int)
	newest := new.(int)
	if oldest != 50 || newest != 51 {
		t.Errorf("Ends() values incorrect, got (%d, %d), expected (50, 52)", oldest, newest)
	}
}

func TestLatest(t *testing.T) {
	l := MakeQueueHistory(10)

	for i := 50; i < 52; i++ {
		l.Add(i)
	}

	new, err := l.Latest()
	if err != nil {
		t.Errorf("got error on Ends(): %s", err)
	}

	newest := new.(int)
	if newest != 51 {
		t.Errorf("Latest() value incorrect, got %d, expected 52", newest)
	}
}
