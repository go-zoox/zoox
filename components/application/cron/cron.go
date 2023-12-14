package cron

import (
	"fmt"

	gocron "github.com/go-zoox/cron"
)

// Cron ...
type Cron interface {
	AddJob(id string, spec string, job func() error) (err error)
	RemoveJob(id string) error
	HasJob(id string) bool
	ClearJobs() error
	AddSecondlyJob(id string, cmd func() error) (err error)
	AddMinutelyJob(id string, cmd func() error) (err error)
	AddHourlyJob(id string, cmd func() error) (err error)
	AddDailyJob(id string, cmd func() error) (err error)
	AddWeeklyJob(id string, cmd func() error) (err error)
	AddMonthlyJob(id string, cmd func() error) (err error)
	AddYearlyJob(id string, cmd func() error) (err error)
}

type cron struct {
	isStarted bool
	core      *gocron.Cron
}

// New creates a cron.
func New() Cron {
	core, err := gocron.New()
	if err != nil {
		panic(err)
	}

	return &cron{
		core: core,
	}
}

// AddJob ...
func (c *cron) AddJob(id string, spec string, job func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddJob(id, spec, job)
}

// RemoveJob ...
func (c *cron) RemoveJob(id string) error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.RemoveJob(id)
}

// HasJob
func (c *cron) HasJob(id string) bool {
	if !c.isStarted {
		return false
	}

	return c.core.HasJob(id)
}

// ClearJobs clears all jobs.
func (c *cron) ClearJobs() error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.ClearJobs()
}

// AddSecondlyJob adds a schedule job run in every second.
func (c *cron) AddSecondlyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddSecondlyJob(id, cmd)
}

// AddMinutelyJob adds a schedule job run in every minute.
func (c *cron) AddMinutelyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMinutelyJob(id, cmd)
}

// AddHourlyJob adds a schedule job run in every hour.
func (c *cron) AddHourlyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddHourlyJob(id, cmd)
}

// AddDailyJob adds a schedule job run in every day.
func (c *cron) AddDailyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddDailyJob(id, cmd)
}

// AddWeeklyJob adds a schedule job run in every week.
func (c *cron) AddWeeklyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddWeeklyJob(id, cmd)
}

// AddMonthlyJob adds a schedule job run in every month.
func (c *cron) AddMonthlyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMonthlyJob(id, cmd)
}

// AddYearlyJob adds a schedule job run in every year.
func (c *cron) AddYearlyJob(id string, cmd func() error) (err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddYearlyJob(id, cmd)
}
