package app

func (a *App[EnvInterface]) AddCronJob(schedule, jobName string, job func()) {
	a.container.NewCron()

	if err := a.container.AddCronJob(schedule, jobName, job); err != nil {
		a.container.Logger.Error("error adding cron job", err)
	}
}
