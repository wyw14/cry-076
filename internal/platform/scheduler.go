package platform

import (
	"context"
	"sync"
	"time"
)

type ScheduledJob struct {
	Name  string
	Every time.Duration
	Run   func(context.Context) error
}
type DraftMaintenanceScheduler struct {
	jobs []ScheduledJob
	wg   sync.WaitGroup
}

func (s *DraftMaintenanceScheduler) Register(job ScheduledJob) {
	if job.Every <= 0 || job.Run == nil {
		panic("invalid scheduled job")
	}
	s.jobs = append(s.jobs, job)
}
func (s *DraftMaintenanceScheduler) Start(ctx context.Context) {
	for _, job := range s.jobs {
		job := job
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(job.Every)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					_ = job.Run(ctx)
				}
			}
		}()
	}
}
func (s *DraftMaintenanceScheduler) Wait() { s.wg.Wait() }
