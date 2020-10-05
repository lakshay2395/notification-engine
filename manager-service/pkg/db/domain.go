package db

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Status string

const (
	UNPICKED = Status("READY")
	PENDING  = Status("QUEUED")
	DONE     = Status("DONE")
)

type Shard struct {
	Id        uuid.UUID `json:"id" db:"id"`
	BucketId  int64     `json:"bucket_id" db:"bucket_id"`
	CreatedOn time.Time `json:"created_on" db:"created_on"`
}

type Task struct {
	Id               uuid.UUID `json:"id" db:"id"`
	ShardId          uuid.UUID `json:"shard_id" db:"shard_id"`
	Status           Status    `json:"status" db:"status"`
	Source           string    `json:"source" db:"source"`
	Destination      string    `json:"destination" db:"destination"`
	TimeForExecution time.Time `json:"time_for_execution" db:"time_for_execution"`
	Content          string    `json:"content" db:"content"`
}
