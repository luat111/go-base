package pkg

import (
	"go-base/pkg/common"
	"go-base/pkg/logger"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Cronjob struct {
	cron   *cron.Cron
	logger logger.ILogger

	mu sync.RWMutex
}

func NewCron() *Cronjob {
	logger := logger.NewLogger(common.CronPrefix)
	c := &Cronjob{
		cron:   cron.New(),
		logger: logger,
	}

	return c
}

func (c *Cronjob) AddJob(schedule, jobName string, fn func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.cron.AddFunc(schedule, func() {
		start := time.Now()
		c.logger.Info("Executing cron job:", jobName)
		fn()
		c.logger.Info("Job finished:", jobName, "Duration:", time.Since(start))
	})

	return err
}

func (c *Cronjob) Run() {
	c.cron.Start()
}

func (c *Cronjob) Stop() {
	c.cron.Stop()
}
