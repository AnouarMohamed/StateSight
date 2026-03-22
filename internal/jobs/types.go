package jobs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AnouarMohamed/StateSight/pkg/model"
)

type Message struct {
	JobID         string            `json:"job_id"`
	JobType       string            `json:"job_type"`
	ApplicationID string            `json:"application_id,omitempty"`
	Payload       map[string]any    `json:"payload,omitempty"`
	EnqueuedAt    time.Time         `json:"enqueued_at"`
	Headers       map[string]string `json:"headers,omitempty"`
}

func (m Message) Validate() error {
	switch m.JobType {
	case model.JobTypeAnalyzeApplication, model.JobTypeIngestGitHubEvent:
		return nil
	default:
		return fmt.Errorf("unsupported job type: %s", m.JobType)
	}
}

func (m Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func UnmarshalMessage(raw []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}
