package notifications

import (
	"fmt"
	"log"
	"time"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	TaskAssignedNotification     NotificationType = "task_assigned"
	TaskUpdatedNotification      NotificationType = "task_updated"
	ProjectInvitationNotification NotificationType = "project_invitation"
	TaskCompletedNotification    NotificationType = "task_completed"
	TaskOverdueNotification      NotificationType = "task_overdue"
)

// Notification represents a notification to be sent
type Notification struct {
	ID          string           `json:"id"`
	Type        NotificationType `json:"type"`
	RecipientID uint             `json:"recipient_id"`
	SenderID    uint             `json:"sender_id"`
	ProjectID   uint             `json:"project_id"`
	TaskID      *uint            `json:"task_id,omitempty"`
	Title       string           `json:"title"`
	Message     string           `json:"message"`
	Data        interface{}      `json:"data"`
	CreatedAt   time.Time        `json:"created_at"`
	SentAt      *time.Time       `json:"sent_at,omitempty"`
	Status      string           `json:"status"` // "pending", "sent", "failed"
}

// NotificationService handles background notification processing
type NotificationService struct {
	// Channel for queuing notifications
	notificationQueue chan *Notification
	
	// Channel for controlling service shutdown
	shutdownChan chan bool
	
	// Number of worker goroutines
	workers int
}

// NewNotificationService creates a new notification service
func NewNotificationService(workers int) *NotificationService {
	return &NotificationService{
		notificationQueue: make(chan *Notification, 1000), // Buffered channel for 1000 notifications
		shutdownChan:      make(chan bool),
		workers:           workers,
	}
}

// Start initializes the notification service and starts worker goroutines
func (ns *NotificationService) Start() {
	log.Printf("📧 Starting Notification Service with %d workers", ns.workers)
	
	// Start multiple worker goroutines for concurrent processing
	for i := 0; i < ns.workers; i++ {
		go ns.worker(i + 1)
	}
	
	log.Println("✅ Notification Service started - Background notifications enabled!")
}

// Stop gracefully shuts down the notification service
func (ns *NotificationService) Stop() {
	log.Println("🛑 Stopping Notification Service...")
	close(ns.shutdownChan)
}

// worker is a goroutine that processes notifications from the queue
func (ns *NotificationService) worker(workerID int) {
	log.Printf("👷 Notification Worker %d started", workerID)
	
	for {
		select {
		case notification := <-ns.notificationQueue:
			// Process the notification
			ns.processNotification(notification, workerID)
			
		case <-ns.shutdownChan:
			log.Printf("👷 Notification Worker %d shutting down", workerID)
			return
		}
	}
}

// processNotification handles the actual sending of a notification
func (ns *NotificationService) processNotification(notification *Notification, workerID int) {
	log.Printf("📤 Worker %d processing %s notification for user %d", 
		workerID, notification.Type, notification.RecipientID)
	
	// Simulate processing time (in a real system, this would be API calls to email/SMS services)
	time.Sleep(time.Millisecond * 100)
	
	// Simulate different notification channels based on type
	switch notification.Type {
	case TaskAssignedNotification:
		ns.sendTaskAssignedEmail(notification)
	case TaskUpdatedNotification:
		ns.sendTaskUpdatedEmail(notification)
	case ProjectInvitationNotification:
		ns.sendProjectInvitationEmail(notification)
	case TaskCompletedNotification:
		ns.sendTaskCompletedEmail(notification)
	case TaskOverdueNotification:
		ns.sendTaskOverdueEmail(notification)
	default:
		log.Printf("⚠️ Unknown notification type: %s", notification.Type)
	}
	
	// Mark notification as sent
	now := time.Now()
	notification.SentAt = &now
	notification.Status = "sent"
	
	log.Printf("✅ Worker %d completed %s notification for user %d", 
		workerID, notification.Type, notification.RecipientID)
}

// QueueNotification adds a notification to the processing queue
func (ns *NotificationService) QueueNotification(notification *Notification) {
	// Set default values
	notification.CreatedAt = time.Now()
	notification.Status = "pending"
	
	// Generate unique ID
	notification.ID = generateNotificationID()
	
	// Try to add to queue (non-blocking)
	select {
	case ns.notificationQueue <- notification:
		log.Printf("📥 Queued %s notification for user %d", notification.Type, notification.RecipientID)
	default:
		log.Printf("⚠️ Notification queue is full, dropping notification for user %d", notification.RecipientID)
		notification.Status = "failed"
	}
}

// Notification sender methods (simulating email/SMS/push notifications)

func (ns *NotificationService) sendTaskAssignedEmail(notification *Notification) {
	log.Printf("📧 [EMAIL] Task Assigned: %s - To: User %d", notification.Message, notification.RecipientID)
	// In a real system, this would integrate with SendGrid, AWS SES, etc.
}

func (ns *NotificationService) sendTaskUpdatedEmail(notification *Notification) {
	log.Printf("📧 [EMAIL] Task Updated: %s - To: User %d", notification.Message, notification.RecipientID)
}

func (ns *NotificationService) sendProjectInvitationEmail(notification *Notification) {
	log.Printf("📧 [EMAIL] Project Invitation: %s - To: User %d", notification.Message, notification.RecipientID)
}

func (ns *NotificationService) sendTaskCompletedEmail(notification *Notification) {
	log.Printf("📧 [EMAIL] Task Completed: %s - To: User %d", notification.Message, notification.RecipientID)
}

func (ns *NotificationService) sendTaskOverdueEmail(notification *Notification) {
	log.Printf("📧 [EMAIL] Task Overdue: %s - To: User %d", notification.Message, notification.RecipientID)
}

// Helper functions

func generateNotificationID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

// Convenience functions for common notification scenarios

// SendTaskAssignedNotification sends a notification when a task is assigned to a user
func (ns *NotificationService) SendTaskAssignedNotification(recipientID, senderID, projectID, taskID uint, taskTitle string) {
	notification := &Notification{
		Type:        TaskAssignedNotification,
		RecipientID: recipientID,
		SenderID:    senderID,
		ProjectID:   projectID,
		TaskID:      &taskID,
		Title:       "New Task Assigned",
		Message:     fmt.Sprintf("You have been assigned a new task: %s", taskTitle),
		Data: map[string]interface{}{
			"task_id":    taskID,
			"task_title": taskTitle,
			"project_id": projectID,
		},
	}
	ns.QueueNotification(notification)
}

// SendTaskUpdatedNotification sends a notification when a task is updated
func (ns *NotificationService) SendTaskUpdatedNotification(recipientID, senderID, projectID, taskID uint, taskTitle string, changes map[string]interface{}) {
	notification := &Notification{
		Type:        TaskUpdatedNotification,
		RecipientID: recipientID,
		SenderID:    senderID,
		ProjectID:   projectID,
		TaskID:      &taskID,
		Title:       "Task Updated",
		Message:     fmt.Sprintf("Task '%s' has been updated", taskTitle),
		Data: map[string]interface{}{
			"task_id":    taskID,
			"task_title": taskTitle,
			"project_id": projectID,
			"changes":    changes,
		},
	}
	ns.QueueNotification(notification)
}

// SendProjectInvitationNotification sends a notification when a user is invited to a project
func (ns *NotificationService) SendProjectInvitationNotification(recipientID, senderID, projectID uint, projectName, inviterName string) {
	notification := &Notification{
		Type:        ProjectInvitationNotification,
		RecipientID: recipientID,
		SenderID:    senderID,
		ProjectID:   projectID,
		Title:       "Project Invitation",
		Message:     fmt.Sprintf("%s invited you to collaborate on '%s'", inviterName, projectName),
		Data: map[string]interface{}{
			"project_id":   projectID,
			"project_name": projectName,
			"inviter_name": inviterName,
		},
	}
	ns.QueueNotification(notification)
}

// SendTaskCompletedNotification sends a notification when a task is completed
func (ns *NotificationService) SendTaskCompletedNotification(recipientID, senderID, projectID, taskID uint, taskTitle string) {
	notification := &Notification{
		Type:        TaskCompletedNotification,
		RecipientID: recipientID,
		SenderID:    senderID,
		ProjectID:   projectID,
		TaskID:      &taskID,
		Title:       "Task Completed",
		Message:     fmt.Sprintf("Task '%s' has been completed", taskTitle),
		Data: map[string]interface{}{
			"task_id":    taskID,
			"task_title": taskTitle,
			"project_id": projectID,
		},
	}
	ns.QueueNotification(notification)
}
