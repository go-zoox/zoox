package zoox

import (
	"runtime"

	"github.com/go-zoox/jobqueue"
)

// Queue is a simple job queue.
type Queue interface {
	AddJob(job jobqueue.Job) error
	AddJobFunc(task func(), callback func(status int, err error)) error
}

type queue struct {
	isStarted bool
	core      *jobqueue.JobQueue
}

func newQueue() Queue {
	core := jobqueue.New(runtime.NumCPU())

	return &queue{
		core: core,
	}
}

// AddJob ...
func (q *queue) AddJob(job jobqueue.Job) error {
	if !q.isStarted {
		q.core.Start()
	}

	q.core.AddJob(job)
	return nil
}

// AddJobFunc ...
func (q *queue) AddJobFunc(task func(), callback func(status int, err error)) error {
	if !q.isStarted {
		q.core.Start()
	}

	return q.AddJob(jobqueue.NewJob(task, callback))
}
