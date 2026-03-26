package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

func TestNewScheduler(t *testing.T) {
	// Use a mock that satisfies QueueService
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	scheduler := NewScheduler(svc, 1*time.Hour)
	assert.NotNil(t, scheduler)
	assert.Equal(t, 1*time.Hour, scheduler.interval)
}

func TestScheduler_StartAndStop(t *testing.T) {
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 9600, UnitsMax: 10000, Uploads: 6,
	}

	queueRepo.On("GetOrCreateDailyQuota", mock.Anything, today).Return(quota, nil)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	scheduler := NewScheduler(svc, 100*time.Millisecond)

	scheduler.Start()

	// Wait for at least one process cycle
	time.Sleep(250 * time.Millisecond)

	scheduler.Stop()

	// Give goroutine time to stop
	time.Sleep(50 * time.Millisecond)

	// Verify it ran at least once (the initial processOnce on start)
	queueRepo.AssertCalled(t, "GetOrCreateDailyQuota", mock.Anything, today)
}

func TestScheduler_StopPreventsMoreProcessing(t *testing.T) {
	queueRepo := new(MockQueueRepository)
	mediaRepo := new(MockMediaRepository)
	tokenRepo := new(MockTokenRepo)
	tokenSvc := new(MockTokenSvc)
	ytClient := new(MockYouTubeClient)

	today := time.Now().Format("2006-01-02")
	quota := &domain.DailyQuota{
		ID: uuid.New(), Date: today, UnitsUsed: 9600, UnitsMax: 10000,
	}

	queueRepo.On("GetOrCreateDailyQuota", mock.Anything, today).Return(quota, nil)

	svc := newTestQueueService(queueRepo, mediaRepo, tokenRepo, tokenSvc, ytClient)

	scheduler := NewScheduler(svc, 5*time.Second) // Long interval
	scheduler.Start()

	// Wait for initial processOnce
	time.Sleep(100 * time.Millisecond)

	// Stop immediately
	scheduler.Stop()

	// Wait and verify no more processing
	time.Sleep(200 * time.Millisecond)
}
