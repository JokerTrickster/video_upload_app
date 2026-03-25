package service

import (
	"context"
	"time"

	"github.com/JokerTrickster/video-upload-backend/internal/pkg/logger"
)

// Scheduler runs the queue processor on a schedule
type Scheduler struct {
	queueService QueueService
	interval     time.Duration
	stopCh       chan struct{}
}

// NewScheduler creates a new scheduler
func NewScheduler(queueService QueueService, interval time.Duration) *Scheduler {
	return &Scheduler{
		queueService: queueService,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the scheduler loop in a goroutine
func (s *Scheduler) Start() {
	go func() {
		logger.Info("Upload queue scheduler started",
			"interval", s.interval.String())

		// Process immediately on start
		s.processOnce()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.processOnce()
			case <-s.stopCh:
				logger.Info("Upload queue scheduler stopped")
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) processOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.Info("Running scheduled queue processing")
	if err := s.queueService.ProcessQueue(ctx); err != nil {
		logger.Error("Queue processing failed", "error", err)
	}
}
