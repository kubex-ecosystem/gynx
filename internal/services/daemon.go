// Package services provides background service orchestration for the gnyx.
// This enables autonomous operation independent of web interface.
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubex-ecosystem/gnyx/internal/types"

	gl "github.com/kubex-ecosystem/logz"
)

// DaemonService provides autonomous background operations for repository analysis,
// scheduling, notifications, and integration with external tools like lookatni/grompt.
type DaemonService struct {
	config           *types.Config
	notificationSvc  *NotificationService
	schedulerSvc     *SchedulerService
	orchestrationSvc *OrchestrationService

	// Internal state
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex

	// Channels for communication
	analysisRequests  chan types.AnalysisRequest
	notificationQueue chan types.NotificationEvent
	orchestrateQueue  chan types.OrchestrationTask
}

// NewDaemonService creates a new daemon service instance
func NewDaemonService(config *types.Config) *DaemonService {
	ctx, cancel := context.WithCancel(context.Background())

	return &DaemonService{
		config:            config,
		running:           false,
		ctx:               ctx,
		cancel:            cancel,
		analysisRequests:  make(chan types.AnalysisRequest, 100),
		notificationQueue: make(chan types.NotificationEvent, 500),
		orchestrateQueue:  make(chan types.OrchestrationTask, 200),
	}
}

// Start begins autonomous daemon operations
func (d *DaemonService) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return gl.Errorf("daemon service already running")
	}

	gl.Log("info", "Starting GNyx Daemon Service...")

	// Initialize sub-services
	if err := d.initializeServices(); err != nil {
		return gl.Errorf("failed to initialize services: %v", err)
	}

	// Start worker goroutines
	d.startWorkers()

	d.running = true
	gl.Log("info", "GNyx Daemon Service started successfully")

	return nil
}

// Stop gracefully shuts down the daemon service
func (d *DaemonService) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	gl.Log("info", "🛑 Stopping GNyx Daemon Service...")

	// Signal all workers to stop
	d.cancel()

	// Wait for all workers to finish with timeout
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		gl.Log("info", "All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		gl.Log("warn", "Timeout waiting for workers to stop")
	}

	d.running = false
	gl.Log("info", "GNyx Daemon Service stopped")

	return nil
}

// IsRunning returns the current running status
func (d *DaemonService) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// ScheduleAnalysis adds a new analysis request to the queue
func (d *DaemonService) ScheduleAnalysis(req types.AnalysisRequest) error {
	if req.ID == "" {
		req.ID = fmt.Sprintf("analysis_%d", time.Now().UnixNano())
	}
	req.CreatedAt = time.Now()

	select {
	case d.analysisRequests <- req:
		gl.Log("info", fmt.Sprintf("📊 Scheduled analysis: %s for project: %s", req.Type, req.ProjectPath))
		return nil
	case <-d.ctx.Done():
		return gl.Errorf("daemon service is shutting down")
	default:
		return gl.Errorf("analysis queue is full")
	}
}

// SendNotification adds a notification to the queue
func (d *DaemonService) SendNotification(event types.NotificationEvent) error {
	event.CreatedAt = time.Now()

	select {
	case d.notificationQueue <- event:
		gl.Log("info", fmt.Sprintf("🔔 Queued %s notification: %s", event.Type, event.Subject))
		return nil
	case <-d.ctx.Done():
		return gl.Errorf("daemon service is shutting down")
	default:
		return gl.Errorf("notification queue is full")
	}
}

// OrchestrateTool adds an orchestration task to the queue
func (d *DaemonService) OrchestrateTool(task types.OrchestrationTask) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("orchestration_%d", time.Now().UnixNano())
	}
	task.CreatedAt = time.Now()

	select {
	case d.orchestrateQueue <- task:
		gl.Log("info", fmt.Sprintf("🎯 Queued orchestration: %s -> %s", task.Tool, task.Action))
		return nil
	case <-d.ctx.Done():
		return gl.Errorf("daemon service is shutting down")
	default:
		return gl.Errorf("orchestration queue is full")
	}
}

// initializeServices sets up all sub-services
func (d *DaemonService) initializeServices() error {
	// Initialize notification service
	d.notificationSvc = NewNotificationService(
		d.config,
	)

	// Initialize scheduler service
	d.schedulerSvc = NewSchedulerService(
		d.config,
	)

	// Initialize orchestration service
	d.orchestrationSvc = NewOrchestrationService(d.config)

	return nil
}

// startWorkers launches all background worker goroutines
func (d *DaemonService) startWorkers() {
	// Analysis worker
	d.wg.Add(1)
	go d.analysisWorker()

	// Notification worker
	d.wg.Add(1)
	go d.notificationWorker()

	// Orchestration worker
	d.wg.Add(1)
	go d.orchestrationWorker()

	// Health check worker
	d.wg.Add(1)
	go d.healthCheckWorker()

	// Scheduler worker
	d.wg.Add(1)
	go d.schedulerWorker()
}

// analysisWorker processes analysis requests
func (d *DaemonService) analysisWorker() {
	defer d.wg.Done()

	for {
		select {
		case req := <-d.analysisRequests:
			d.processAnalysisRequest(req)
		case <-d.ctx.Done():
			gl.Log("info", "📊 Analysis worker stopped")
			return
		}
	}
}

// notificationWorker processes notification events
func (d *DaemonService) notificationWorker() {
	defer d.wg.Done()

	for {
		select {
		case event := <-d.notificationQueue:
			d.processNotificationEvent(event)
		case <-d.ctx.Done():
			gl.Log("info", "🔔 Notification worker stopped")
			return
		}
	}
}

// orchestrationWorker processes orchestration tasks
func (d *DaemonService) orchestrationWorker() {
	defer d.wg.Done()

	for {
		select {
		case task := <-d.orchestrateQueue:
			d.processOrchestrationTask(task)
		case <-d.ctx.Done():
			gl.Log("info", "🎯 Orchestration worker stopped")
			return
		}
	}
}

// healthCheckWorker performs periodic health checks
func (d *DaemonService) healthCheckWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.performHealthCheck()
		case <-d.ctx.Done():
			gl.Log("info", "❤️ Health check worker stopped")
			return
		}
	}
}

// schedulerWorker handles scheduled tasks
func (d *DaemonService) schedulerWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.checkScheduledTasks()
		case <-d.ctx.Done():
			gl.Log("info", "⏰ Scheduler worker stopped")
			return
		}
	}
}

// processAnalysisRequest handles individual analysis requests
func (d *DaemonService) processAnalysisRequest(req types.AnalysisRequest) {
	gl.Log("info", fmt.Sprintf("🔍 Processing analysis: %s for %s", req.Type, req.ProjectPath))

	// TODO: Implement actual analysis logic
	// This will use the scorecard engine to perform repository analysis

	// Simulate analysis work
	time.Sleep(2 * time.Second)

	// Send completion notification
	d.SendNotification(types.NotificationEvent{
		Type:     "discord",
		Subject:  fmt.Sprintf("Analysis Complete: %s", req.Type),
		Content:  fmt.Sprintf("Repository analysis completed for: %s", req.ProjectPath),
		Priority: "medium",
	})

	gl.Log("info", fmt.Sprintf("Analysis completed: %s", req.ID))
}

// processNotificationEvent handles individual notification events
func (d *DaemonService) processNotificationEvent(event types.NotificationEvent) {
	gl.Log("info", fmt.Sprintf("📤 Sending %s notification: %s", event.Type, event.Subject))

	// TODO: Implement actual notification sending
	// This will use the notification service to send via Discord/WhatsApp/Email

	gl.Log("info", fmt.Sprintf("Notification sent: %s", event.Type))
}

// processOrchestrationTask handles individual orchestration tasks
func (d *DaemonService) processOrchestrationTask(task types.OrchestrationTask) {
	gl.Log("info", fmt.Sprintf("Orchestrating: %s -> %s", task.Tool, task.Action))

	// TODO: Implement actual orchestration logic
	// This will coordinate with lookatni, grompt, and other agents

	gl.Log("info", fmt.Sprintf("Orchestration completed: %s", task.ID))
}

// performHealthCheck performs system health checks
func (d *DaemonService) performHealthCheck() {
	gl.Log("info", "❤️ Performing health check...")

	// TODO: Implement health check logic
	// Check system resources, external service connectivity, etc.

	gl.Log("info", "Health check completed")
}

// checkScheduledTasks checks for and executes scheduled tasks
func (d *DaemonService) checkScheduledTasks() {
	// TODO: Implement scheduled task checking
	// This will check the scheduler service for tasks that need to be executed
}
