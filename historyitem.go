package history

import (
	"time"
)

type HistoryItem interface{}

type HistoryItemWithTime struct {
	t time.Time   // time associated with item
	h HistoryItem // item
}
