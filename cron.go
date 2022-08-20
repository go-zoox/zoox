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
func (c *Cron) AddJob(spec string, job func() error) error {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddJob(spec, job)
	return nil
}

// AddSecondlyJob adds a schedule job run in every second.
func (c *Cron) AddSecondlyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddSecondlyJob(cmd)
}

// AddMinutelyJob adds a schedule job run in every minute.
func (c *Cron) AddMinutelyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddMinutelyJob(cmd)
}

// AddHourlyJob adds a schedule job run in every hour.
func (c *Cron) AddHourlyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddHourlyJob(cmd)
}

// AddDailyJob adds a schedule job run in every day.
func (c *Cron) AddDailyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddDailyJob(cmd)
}

// AddWeeklyJob adds a schedule job run in every week.
func (c *Cron) AddWeeklyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddWeeklyJob(cmd)
}

// AddMonthlyJob adds a schedule job run in every month.
func (c *Cron) AddMonthlyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddMonthlyJob(cmd)
}

// AddYearlyJob adds a schedule job run in every year.
func (c *Cron) AddYearlyJob(cmd func() error) {
	if !c.isStarted {
		c.core.Start()
	}

	c.core.AddYearlyJob(cmd)
}
