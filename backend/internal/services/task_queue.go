package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/alessandrocruz5/scrappd-app/backend/internal/config"
	"github.com/sirupsen/logrus"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

// TaskQueue enqueues background jobs.
type TaskQueue interface {
	EnqueueProcessItem(ctx context.Context, payload ProcessItemTaskPayload) error
}

// ProcessItemTaskPayload represents the task payload for item processing.
type ProcessItemTaskPayload struct {
	ItemID       string `json:"item_id"`
	UserID       string `json:"user_id"`
	OutputFormat string `json:"output_format"`
}

type cloudTasksQueue struct {
	client              *cloudtasks.Client
	projectID           string
	location            string
	queueID             string
	serviceURL          string
	serviceAccountEmail string
	internalSecret      string
	logger              *logrus.Logger
}

// NewCloudTasksQueue creates a Cloud Tasks-backed queue.
func NewCloudTasksQueue(cfg *config.CloudTasksConfig, internalSecret string, logger *logrus.Logger) (TaskQueue, error) {
	if cfg.ProjectID == "" || cfg.Location == "" || cfg.QueueID == "" || cfg.ServiceURL == "" {
		return nil, fmt.Errorf("cloud tasks configuration is incomplete")
	}

	client, err := cloudtasks.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud tasks client: %w", err)
	}

	return &cloudTasksQueue{
		client:              client,
		projectID:           cfg.ProjectID,
		location:            cfg.Location,
		queueID:             cfg.QueueID,
		serviceURL:          strings.TrimRight(cfg.ServiceURL, "/"),
		serviceAccountEmail: cfg.ServiceAccountEmail,
		internalSecret:      internalSecret,
		logger:              logger,
	}, nil
}

func (q *cloudTasksQueue) EnqueueProcessItem(ctx context.Context, payload ProcessItemTaskPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/queues/%s", q.projectID, q.location, q.queueID)
	task := &taskspb.Task{
		MessageType: &taskspb.Task_HttpRequest{
			HttpRequest: &taskspb.HttpRequest{
				HttpMethod: taskspb.HttpMethod_POST,
				Url:        fmt.Sprintf("%s/internal/items/process", q.serviceURL),
				Headers: map[string]string{
					"Content-Type":      "application/json",
					"X-Internal-Secret": q.internalSecret,
				},
				Body: body,
			},
		},
	}

	if q.serviceAccountEmail != "" {
		task.GetHttpRequest().AuthorizationHeader = &taskspb.HttpRequest_OidcToken{
			OidcToken: &taskspb.OidcToken{
				ServiceAccountEmail: q.serviceAccountEmail,
			},
		}
	}

	_, err = q.client.CreateTask(ctx, &taskspb.CreateTaskRequest{
		Parent: parent,
		Task:   task,
	})
	if err != nil {
		q.logger.WithError(err).Error("Failed to create cloud task")
		return fmt.Errorf("failed to create cloud task: %w", err)
	}

	return nil
}
