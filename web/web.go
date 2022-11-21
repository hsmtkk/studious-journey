package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/genproto/googleapis/api/metric"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	port, err := getPort()
	if err != nil {
		log.Fatal(err)
	}

	hdl := newHandler()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hdl.index)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

type handler struct{}

func newHandler() *handler {
	return &handler{}
}

// Handler
func (h *handler) index(ectx echo.Context) error {
	ctx := ectx.Request().Context()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.example.com", nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext failed; %w", err)
	}
	before := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do failed; %w", err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(before)
	if err := h.recordMetrics(ctx, elapsed); err != nil {
		return err
	}
	return ectx.String(http.StatusOK, "Hello, World!")
}

func (h *handler) recordMetrics(ctx context.Context, elapsed time.Duration) error {
	projectID, err := h.projectID(ctx)
	if err != nil {
		return err
	}
	clt, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("monitoring.NewMetricClient failed; %w", err)
	}
	defer clt.Close()
	now := timestamppb.Now()
	req := monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + projectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metric.Metric{
				Type: "custom.googleapis.com/foo/bar",
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: now,
					EndTime:   now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: elapsed.Milliseconds(),
					},
				},
			}},
		}},
	}
	if err := clt.CreateTimeSeries(ctx, &req); err != nil {
		return fmt.Errorf("monitoring.MetricClient.CreateTimeSeries failed; %w", err)
	}
	return nil
}

func (h *handler) projectID(ctx context.Context) (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", fmt.Errorf("google.FindDefaultCredentials failed; %w", err)
	}
	return credentials.ProjectID, nil
}
