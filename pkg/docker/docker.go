package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"io"
	"os"
	"time"
)

const (
	ApiVersion = "1.30"

	ContainerStateRunning    = "running"
	ContainerStateCreated    = "created"
	ContainerStateExited     = "exited"
	ContainerStateRestarting = "restarting"
	ContainerStatePaused     = "paused"
	ContainerStateDead       = "dead"
)

type Docker struct {
	client *client.Client
	ctx    context.Context
}

func New() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(ApiVersion))
	if err != nil {
		return nil, err
	}

	return &Docker{
		client: cli,
		ctx:    context.Background(),
	}, nil
}

func (docker *Docker) IsImageBuilt(image string) (bool, error) {
	list, err := docker.client.ImageList(docker.ctx, types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].RepoTags {
			if list[i].RepoTags[j] == image {
				return true, nil
			}
		}
	}

	return false, nil
}

func (docker *Docker) IsContainerCreated(container string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+container {
				return true, nil
			}
		}
	}

	return false, nil
}

func (docker *Docker) IsContainerStarted(container string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+container {
				if list[i].State == ContainerStateRunning {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (docker *Docker) IsContainerStopped(container string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+container {
				if list[i].State == ContainerStateRunning {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func (docker *Docker) BuildImage(name, from string) error {
	dockerfile, err := dockerfileParse(from)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	writer := tar.NewWriter(buffer)
	header := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
	}
	options := types.ImageBuildOptions{
		Tags:   []string{name},
		Remove: true,
	}

	err = writer.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(dockerfile))
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	response, err := docker.client.ImageBuild(docker.ctx, buffer, options)
	if err != nil {
		return err
	}

	termFd, isTerm := term.GetFdInfo(os.Stdout)
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	_, _, err = docker.client.ImageInspectWithRaw(docker.ctx, name)
	if err != nil {
		return errors.New("image didn't built successfully")
	}

	return nil
}

func (docker *Docker) CreateContainer(name *naming.Naming) error {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: name.SourceDir,
				Target: naming.ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: name.BuildDir,
				Target: naming.ContainerBuildDir,
			}, {
				Type:   mount.TypeBind,
				Source: name.CacheDir,
				Target: naming.ContainerCacheDir,
			}, {
				Type:   mount.TypeBind,
				Source: name.ArchiveDir,
				Target: naming.ContainerArchiveDir,
			},
		},
	}
	config := &container.Config{
		Image: name.Image,
	}

	// mkdir
	for _, mnt := range hostConfig.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	_, err := docker.client.ContainerCreate(docker.ctx, config, hostConfig, nil, name.Container)
	if err != nil {
		return err
	}

	return nil
}

func (docker *Docker) StartContainer(container string) error {
	options := types.ContainerStartOptions{}

	return docker.client.ContainerStart(docker.ctx, container, options)
}

func (docker *Docker) StopContainer(container string) error {
	timeout := time.Second

	return docker.client.ContainerStop(docker.ctx, container, &timeout)
}

func (docker *Docker) RemoveContainer(container string) error {
	options := types.ContainerRemoveOptions{}

	return docker.client.ContainerRemove(docker.ctx, container, options)
}

func (docker *Docker) ExecContainer(container string, cmd ...string) error {
	config := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	check := types.ExecStartCheck{
		Tty:    true,
		Detach: false,
	}

	response, err := docker.client.ContainerExecCreate(docker.ctx, container, config)
	if err != nil {
		return err
	}

	hijack, err := docker.client.ContainerExecAttach(docker.ctx, response.ID, check)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, hijack.Reader)
	if err != nil {
		return err
	}

	hijack.Close()

	inspect, err := docker.client.ContainerExecInspect(docker.ctx, response.ID)
	if err != nil {
		return err
	}

	if inspect.ExitCode != 0 {
		return errors.New("command exited with non-zero status")
	}

	return nil
}
