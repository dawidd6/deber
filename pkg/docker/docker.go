package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"io"
	"os"
	"path/filepath"
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

	ContainerRepoDir   = "/repo"
	ContainerBuildDir  = "/build"
	ContainerSourceDir = "/build/source"
	ContainerCacheDir  = "/var/cache/apt"
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
				if list[i].State != ContainerStateRunning {
					return true, nil
				}
			}
		}
	}

	return false, nil
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

func HostCacheDir(image string) string {
	return fmt.Sprintf(
		"/tmp/%s",
		image,
	)
}

func HostSourceDir() string {
	return os.Getenv("PWD")
}

func HostBuildDir(container string) string {
	return fmt.Sprintf(
		"%s/../%s",
		HostSourceDir(),
		container,
	)
}

func SourceTarball(tarball string) string {
	return fmt.Sprintf(
		"%s/../%s",
		HostSourceDir(),
		tarball,
	)
}

func TargetTarball(container, tarball string) string {
	return fmt.Sprintf(
		"%s/%s",
		HostBuildDir(container),
		tarball,
	)
}

func (docker *Docker) CreateContainer(name, image, repo, tarball string) error {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: HostCacheDir(image),
				Target: ContainerCacheDir,
			}, {
				Type:   mount.TypeBind,
				Source: HostSourceDir(),
				Target: ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: HostBuildDir(name),
				Target: ContainerBuildDir,
			},
		},
	}
	config := &container.Config{
		Image: image,
	}

	// repo
	if repo != "" {
		repo, err := filepath.Abs(repo)
		if err != nil {
			return err
		}
		mnt := mount.Mount{
			Type:   mount.TypeBind,
			Source: repo,
			Target: ContainerRepoDir,
		}
		hostConfig.Mounts = append(hostConfig.Mounts, mnt)
	}

	// mkdir
	for _, mnt := range hostConfig.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// tarball
	if tarball != "" {
		err := os.Rename(SourceTarball(tarball), TargetTarball(name, tarball))
		if err != nil {
			return err
		}
	}

	_, err := docker.client.ContainerCreate(docker.ctx, config, hostConfig, nil, name)
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

func (docker *Docker) DisconnectAllNetworks(container string) ([]string, error) {
	networks := make([]string, 0)

	json, err := docker.client.ContainerInspect(docker.ctx, container)
	if err != nil {
		return nil, err
	}

	for _, net := range json.NetworkSettings.Networks {
		err := docker.client.NetworkDisconnect(docker.ctx, net.NetworkID, container, false)
		if err != nil {
			return nil, err
		}

		networks = append(networks, net.NetworkID)
	}

	return networks, nil
}

func (docker *Docker) ConnectNetworks(container string, networks []string) error {
	for _, net := range networks {
		err := docker.client.NetworkConnect(docker.ctx, net, container, nil)
		if err != nil {
			return err
		}
	}

	return nil
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
