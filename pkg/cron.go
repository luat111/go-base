package pkg

import (
	"go-base/pkg/container"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Crontab struct {
	cron      *cron.Cron
	container *container.Container

	mu sync.RWMutex
}

func NewCron(ctn *container.Container) *Crontab {
	c := &Crontab{
		cron:      cron.New(),
		container: ctn,
	}

	return c
}

func (c *Crontab) AddJob(schedule, jobName string, fn func()) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.cron.AddFunc(schedule, func() {
		start := time.Now()
		c.container.Logger.Info("Executing cron job:", jobName)
		fn()
		c.container.Logger.Info("Job finished:", jobName, "Duration:", time.Since(start))
	})

	return err
}

func (c *Crontab) Run() {
	c.cron.Start()
}

func (c *Crontab) Stop() {
	c.cron.Stop()
}
