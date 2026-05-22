package httpclient

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	client *http.Client
}

const (
	timeout = 5 * time.Second
)

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: timeout},
	}
}

const (
	ErrCreatingRequest = "error creating HTTP request: %w"
	ErrMakingRequest   = "error making HTTP request: %w"
)

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	log := logger.From(ctx)
	log.Debug("Making HTTP request", "method", req.Method, "url", req.URL.String())

	traceCtx, span := otel.Tracer("Online-queue-management-system/http-client").Start(
		req.Context(),
		req.Method+" "+req.URL.Host+req.URL.Path,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.request.method", req.Method),
			attribute.String("url.full", req.URL.String()),
			attribute.String("server.address", req.URL.Host),
		),
	)
	defer span.End()

	req = req.WithContext(traceCtx)
	otel.GetTextMapPropagator().Inject(traceCtx, propagation.HeaderCarrier(req.Header))
	resp, err := c.client.Do(req)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Error("Error making HTTP request", "method", req.Method, "url", req.URL.String(), "error", err)
		return nil, fmt.Errorf(ErrMakingRequest, err)
	}
	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))

	if resp.Body == nil {
		span.SetStatus(codes.Error, "response body is nil")
		log.Error("Received HTTP response with nil body", "method", req.Method, "url", req.URL.String())
		return nil, fmt.Errorf("received HTTP response with nil body on method %s to url %s", req.Method, req.URL.String())
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		log.Error("Received non-successful HTTP status code", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Error("Error closing response body", "method", req.Method, "url", req.URL.String(), "error", err)
			}
		}()
		return nil, fmt.Errorf("received non-successful HTTP status code on method %s to url %s: %d", req.Method, req.URL.String(), resp.StatusCode)
	}

	log.Debug("HTTP request completed", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode)
	return resp, nil
}

func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	log := logger.From(ctx)
	log.Debug("Creating GET request", "url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Error("Error creating GET request", "url", url, "error", err)
		return nil, fmt.Errorf(ErrCreatingRequest, err)
	}
	log.Debug("GET request created successfully", "url", url)
	return c.Do(ctx, req)
}

func (c *Client) Post(ctx context.Context, url, bodyType string, body io.Reader) (*http.Response, error) {
	log := logger.From(ctx)
	log.Debug("Creating POST request", "url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)

	if err != nil {
		log.Error("Error creating POST request", "url", url, "error", err)
		return nil, fmt.Errorf(ErrCreatingRequest, err)
	}

	req.Header.Set("Content-Type", bodyType)
	log.Debug("POST request created successfully", "url", url)
	return c.Do(ctx, req)
}

func (c *Client) Put(ctx context.Context, url, bodyType string, body io.Reader) (*http.Response, error) {
	log := logger.From(ctx)
	log.Debug("Creating PUT request", "url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)

	if err != nil {
		log.Error("Error creating PUT request", "url", url, "error", err)
		return nil, fmt.Errorf(ErrCreatingRequest, err)
	}
	req.Header.Set("Content-Type", bodyType)
	log.Debug("PUT request created successfully", "url", url)
	return c.Do(ctx, req)
}

func (c *Client) Delete(ctx context.Context, url string) (*http.Response, error) {
	log := logger.From(ctx)
	log.Debug("Creating DELETE request", "url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		log.Error("Error creating DELETE request", "url", url, "error", err)
		return nil, fmt.Errorf(ErrCreatingRequest, err)
	}
	log.Debug("DELETE request created successfully", "url", url)
	return c.Do(ctx, req)
}
