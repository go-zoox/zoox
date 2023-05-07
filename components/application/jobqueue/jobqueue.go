package jobqueue

import (
	"runtime"

	jq "github.com/go-zoox/jobqueue"
)

// JobQueue is a simple job queue.
type JobQueue interface {
	AddJob(job jq.Job) error
	AddJobFunc(task func(), callback func(status int, err error)) error
}

type jobqueue struct {
	isStarted bool
	core      *jq.JobQueue
}

func New() JobQueue {
	core := jq.New(runtime.NumCPU())

	return &jobqueue{
		core: core,
	}
}

// AddJob ...
func (q *jobqueue) AddJob(job jq.Job) error {
	if !q.isStarted {
		q.core.Start()
	}

	q.core.AddJob(job)
	return nil
}

// AddJobFunc ...
func (q *jobqueue) AddJobFunc(task func(), callback func(status int, err error)) error {
	if !q.isStarted {
		q.core.Start()
	}

	return q.AddJob(jq.NewJob(task, callback))
}
