package job

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type JobStatus int

const (
	StatusPending JobStatus = iota
	StatusRunning
	StatusDone
	StatusFailed
	StatusCancelled
)

type Job struct {
	ID         string
	InputPath  string
	OutputPath string // set after EncodingConfig snapshot
	Status     JobStatus
	Progress   float64
	Error      string

	cancelMu sync.Mutex
	cancel   context.CancelFunc
}

func NewJob(input string) *Job {
	return &Job{ID: uuid.NewString(), InputPath: input, Status: StatusPending}
}

func (j *Job) setCancel(c context.CancelFunc) {
	j.cancelMu.Lock()
	defer j.cancelMu.Unlock()
	j.cancel = c
}

func (j *Job) Cancel() {
	j.cancelMu.Lock()
	c := j.cancel
	j.cancelMu.Unlock()
	if c != nil {
		c()
	}
}
