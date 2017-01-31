package history

import (
	"time"
)

type HistoryItem interface{}

type HistoryItemWithTime struct {
	Time time.Time   // time associated with item
	Item HistoryItem // item
}
