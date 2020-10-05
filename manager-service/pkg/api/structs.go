package api

import uuid "github.com/satori/go.uuid"

type TriggerShardRequest struct {
	Shards []uuid.UUID `json:"shards"`
}
