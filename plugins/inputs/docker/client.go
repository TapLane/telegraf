package docker

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
)

var (
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

// StatsResponseReader wraps the raw container stats body. The upstream
// client.ContainerStatsResult no longer carries the daemon OS type, so
// callers must decode the body to determine it.
type StatsResponseReader struct {
	Body io.ReadCloser
}

type dockerClient interface {
	// Info retrieves system-wide information about the Docker server.
	Info(ctx context.Context) (system.Info, error)
	// ContainerList retrieves a list of containers based on the specified options.
	ContainerList(ctx context.Context, options client.ContainerListOptions) ([]container.Summary, error)
	// ContainerStats retrieves real-time statistics for a specific container.
	ContainerStats(ctx context.Context, containerID string, stream bool) (StatsResponseReader, error)
	// ContainerInspect retrieves detailed information about a specific container.
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	// ServiceList retrieves a list of services based on the specified options.
	ServiceList(ctx context.Context, options client.ServiceListOptions) ([]swarm.Service, error)
	// TaskList retrieves a list of tasks based on the specified options.
	TaskList(ctx context.Context, options client.TaskListOptions) ([]swarm.Task, error)
	// NodeList retrieves a list of nodes based on the specified options.
	NodeList(ctx context.Context, options client.NodeListOptions) ([]swarm.Node, error)
	// DiskUsage retrieves disk usage information.
	DiskUsage(ctx context.Context, options client.DiskUsageOptions) (client.DiskUsageResult, error)
	// ClientVersion retrieves the version of the Docker client.
	ClientVersion() string
	// Close releases any resources held by the client.
	Close() error
}

func newEnvClient() (dockerClient, error) {
	dockerClient, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}
	return &socketClient{dockerClient}, nil
}

func newClient(host string, tlsConfig *tls.Config) (dockerClient, error) {
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	httpClient := &http.Client{Transport: transport}

	dockerClient, err := client.New(
		client.WithHTTPHeaders(defaultHeaders),
		client.WithHTTPClient(httpClient),
		client.WithHost(host))
	if err != nil {
		return nil, err
	}

	return &socketClient{dockerClient}, nil
}

type socketClient struct {
	client *client.Client
}

// Info retrieves system-wide information about the Docker server.
func (c *socketClient) Info(ctx context.Context) (system.Info, error) {
	result, err := c.client.Info(ctx, client.InfoOptions{})
	if err != nil {
		return system.Info{}, err
	}
	return result.Info, nil
}

// ContainerList retrieves a list of containers based on the specified options.
func (c *socketClient) ContainerList(ctx context.Context, options client.ContainerListOptions) ([]container.Summary, error) {
	result, err := c.client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// ContainerStats retrieves real-time statistics for a specific container.
func (c *socketClient) ContainerStats(ctx context.Context, containerID string, stream bool) (StatsResponseReader, error) {
	result, err := c.client.ContainerStats(ctx, containerID, client.ContainerStatsOptions{
		Stream:                stream,
		IncludePreviousSample: !stream,
	})
	if err != nil {
		return StatsResponseReader{}, err
	}
	return StatsResponseReader{Body: result.Body}, nil
}

// ContainerInspect retrieves detailed information about a specific container.
func (c *socketClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	result, err := c.client.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return container.InspectResponse{}, err
	}
	return result.Container, nil
}

// ServiceList retrieves a list of services based on the specified options.
func (c *socketClient) ServiceList(ctx context.Context, options client.ServiceListOptions) ([]swarm.Service, error) {
	result, err := c.client.ServiceList(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// TaskList retrieves a list of tasks based on the specified options.
func (c *socketClient) TaskList(ctx context.Context, options client.TaskListOptions) ([]swarm.Task, error) {
	result, err := c.client.TaskList(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// NodeList retrieves a list of nodes based on the specified options.
func (c *socketClient) NodeList(ctx context.Context, options client.NodeListOptions) ([]swarm.Node, error) {
	result, err := c.client.NodeList(ctx, options)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// DiskUsage retrieves disk usage information.
func (c *socketClient) DiskUsage(ctx context.Context, options client.DiskUsageOptions) (client.DiskUsageResult, error) {
	return c.client.DiskUsage(ctx, options)
}

// ClientVersion retrieves the version of the Docker client.
func (c *socketClient) ClientVersion() string {
	return c.client.ClientVersion()
}

// Close releases any resources held by the client.
func (c *socketClient) Close() error {
	return c.client.Close()
}
