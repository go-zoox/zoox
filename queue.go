package zoox

import (
	"runtime"

	"github.com/go-zoox/jobqueue"
)

// Queue is a simple job queue.
type Queue struct {
	isStarted bool
	core      *jobqueue.JobQueue
}

func newQueue() *Queue {
	core := jobqueue.New(runtime.NumCPU())

	return &Queue{
		core: core,
	}
}

// AddJob ...
func (q *Queue) AddJob(job jobqueue.Job) error {
	if !q.isStarted {
		q.core.Start()
	}

	q.core.AddJob(job)
	return nil
}

// AddJobFunc ...
func (q *Queue) AddJobFunc(task func(), callback func(status int, err error)) error {
	if !q.isStarted {
		q.core.Start()
	}

	return q.AddJob(jobqueue.NewJob(task, callback))
}
