package zoox

import (
	"fmt"

	"github.com/go-zoox/cron"
)

// Cron ...
type Cron struct {
	isStarted bool
	core      *cron.Cron
}

func newCron() *Cron {
	core, err := cron.New()
	if err != nil {
		panic(err)
	}

	return &Cron{
		core: core,
	}
}

// AddJob ...
func (c *Cron) AddJob(name string, spec string, job func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddJob(name, spec, job)
}

// RemoveJob ...
func (c *Cron) RemoveJob(id int) error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.RemoveJob(id)
}

// ClearJobs clears all jobs.
func (c *Cron) ClearJobs() error {
	if !c.isStarted {
		return fmt.Errorf("cron job is not started yet")
	}

	return c.core.ClearJobs()
}

// AddSecondlyJob adds a schedule job run in every second.
func (c *Cron) AddSecondlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddSecondlyJob(name, cmd)
}

// AddMinutelyJob adds a schedule job run in every minute.
func (c *Cron) AddMinutelyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMinutelyJob(name, cmd)
}

// AddHourlyJob adds a schedule job run in every hour.
func (c *Cron) AddHourlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddHourlyJob(name, cmd)
}

// AddDailyJob adds a schedule job run in every day.
func (c *Cron) AddDailyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddDailyJob(name, cmd)
}

// AddWeeklyJob adds a schedule job run in every week.
func (c *Cron) AddWeeklyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddWeeklyJob(name, cmd)
}

// AddMonthlyJob adds a schedule job run in every month.
func (c *Cron) AddMonthlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddMonthlyJob(name, cmd)
}

// AddYearlyJob adds a schedule job run in every year.
func (c *Cron) AddYearlyJob(name string, cmd func() error) (id int, err error) {
	if !c.isStarted {
		c.core.Start()
	}

	return c.core.AddYearlyJob(name, cmd)
}
