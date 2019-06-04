// Package docker wraps Docker Go SDK for internal usage in deber.
package docker

import (
	"context"
	"github.com/docker/docker/client"
)

const (
	// APIVersion constant is the minimum supported version of Docker Engine API
	APIVersion = "1.30"
)

// Docker struct holds Docker client and a context for it.
type Docker struct {
	client *client.Client
	ctx    context.Context
}

// New function creates fresh Docker struct and connects to Docker Engine.
func New() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(APIVersion))
	if err != nil {
		return nil, err
	}

	return &Docker{
		client: cli,
		ctx:    context.Background(),
	}, nil
}
