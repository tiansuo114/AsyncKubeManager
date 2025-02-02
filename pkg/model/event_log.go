package model

type EventLog struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	EventType EventType `gorm:"not null" json:"event_type"`
	Operation string    `gorm:"not null" json:"operation"`
	CreatedAt int64     `gorm:"autoCreateTime:milli; not null" json:"created_at"`
	Creator   string    `gorm:"not null" json:"creator"`
}

type EventType int64

const (
	EventTypeCreateVM   EventType = 1
	EventTypeDeleteVM   EventType = 2
	EventTypeUpdateVM   EventType = 3
	EventTypeCreateDisk EventType = 4
	EventTypeDeleteDisk EventType = 5
	EventTypeUpdateDisk EventType = 6
)

func (EventLog) TableName() string {
	return "event_logs"
}
