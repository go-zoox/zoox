package zoox

import (
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
func (c *Cron) AddJob(name string, spec string, job func() error) error {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddJob(name, spec, job)
	return nil
}

// AddSecondlyJob adds a schedule job run in every second.
func (c *Cron) AddSecondlyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddSecondlyJob(name, cmd)
}

// AddMinutelyJob adds a schedule job run in every minute.
func (c *Cron) AddMinutelyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddMinutelyJob(name, cmd)
}

// AddHourlyJob adds a schedule job run in every hour.
func (c *Cron) AddHourlyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddHourlyJob(name, cmd)
}

// AddDailyJob adds a schedule job run in every day.
func (c *Cron) AddDailyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddDailyJob(name, cmd)
}

// AddWeeklyJob adds a schedule job run in every week.
func (c *Cron) AddWeeklyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddWeeklyJob(name, cmd)
}

// AddMonthlyJob adds a schedule job run in every month.
func (c *Cron) AddMonthlyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddMonthlyJob(name, cmd)
}

// AddYearlyJob adds a schedule job run in every year.
func (c *Cron) AddYearlyJob(name string, cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddYearlyJob(name, cmd)
}
