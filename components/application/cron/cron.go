package cron

import (
	"fmt"

	gocron "github.com/go-zoox/cron"
)

// Cron ...
type Cron interface {
	AddJob(name string, spec string, job func() error) (id int, err error)
	RemoveJob(id int) error
	ClearJobs() error
	AddSecondlyJob(name string, cmd func() error) (id int, err error)
	AddMinutelyJob(name string, cmd func() error) (id int, err error)
	AddHourlyJob(name string, cmd func() error) (id int, err error)
	AddDailyJob(name string, cmd func() error) (id int, err error)
	AddWeeklyJob(name string, cmd func() error) (id int, err error)
	AddMonthlyJob(name string, cmd func() error) (id int, err error)
	AddYearlyJob(name string, cmd func() error) (id int, err error)
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
func (c *cron) AddJob(name string, spec string, job func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddJob(name, spec, job)
}

// RemoveJob ...
func (c *cron) RemoveJob(id int) error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.RemoveJob(id)
}

// ClearJobs clears all jobs.
func (c *cron) ClearJobs() error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.ClearJobs()
}

// AddSecondlyJob adds a schedule job run in every second.
func (c *cron) AddSecondlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddSecondlyJob(name, cmd)
}

// AddMinutelyJob adds a schedule job run in every minute.
func (c *cron) AddMinutelyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMinutelyJob(name, cmd)
}

// AddHourlyJob adds a schedule job run in every hour.
func (c *cron) AddHourlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddHourlyJob(name, cmd)
}

// AddDailyJob adds a schedule job run in every day.
func (c *cron) AddDailyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddDailyJob(name, cmd)
}

// AddWeeklyJob adds a schedule job run in every week.
func (c *cron) AddWeeklyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddWeeklyJob(name, cmd)
}

// AddMonthlyJob adds a schedule job run in every month.
func (c *cron) AddMonthlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMonthlyJob(name, cmd)
}

// AddYearlyJob adds a schedule job run in every year.
func (c *cron) AddYearlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddYearlyJob(name, cmd)
}
