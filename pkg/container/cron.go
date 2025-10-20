package container

import "go-base/pkg"

func (c *Container) NewCron() *pkg.Cronjob {
	if c.cron == nil {
		c.cron = pkg.NewCron()
	}

	return c.cron
}

func (c *Container) StartCron() {
	if c.cron != nil {
		c.cron.Run()
	}
}

func (c *Container) StopCron() {
	if c.cron != nil {
		c.cron.Stop()
	}
}

func (a *Container) AddCronJob(schedule, jobName string, job func()) error {
	return a.cron.AddJob(schedule, jobName, job)
}
