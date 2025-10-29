package app

func (a *App[EnvInterface]) AddCronJob(schedule, jobName string, job func()) {
	a.container.NewCron()

	if err := a.container.AddCronJob(schedule, jobName, job); err != nil {
		a.logger.Error("error adding cron job", err)
	}
}

func (a *App[EnvInterface]) StartCron() {
	a.container.StartCron()
}

func (a *App[EnvInterface]) StopCron() {
	a.container.StopCron()
}
