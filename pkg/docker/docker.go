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
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	ContainerStateRunning    = "running"
	ContainerStateCreated    = "created"
	ContainerStateExited     = "exited"
	ContainerStateRestarting = "restarting"
	ContainerStatePaused     = "paused"
	ContainerStateDead       = "dead"

	ContainerRepoDir     = "/repo"
	ContainerBuildDir    = "/build"
	ContainerSourceDir   = "/build/source"
	ContainerArchivesDir = "/var/cache/apt/archives"
)

type Docker struct {
	client  *client.Client
	ctx     context.Context
	verbose bool
	writer  io.Writer
	buffer  *bytes.Buffer
}

func New(verbose bool) (*Docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	var (
		writer io.Writer
		buffer *bytes.Buffer
	)

	if verbose {
		writer = os.Stdout
	} else {
		buffer = new(bytes.Buffer)
		writer = buffer
	}

	return &Docker{
		client:  cli,
		ctx:     context.Background(),
		verbose: verbose,
		writer:  writer,
		buffer:  buffer,
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

	_, err = io.Copy(docker.writer, response.Body)
	if err != nil {
		return err
	}

	err = response.Body.Close()
	if err != nil {
		return err
	}

	_, _, err = docker.client.ImageInspectWithRaw(docker.ctx, name)
	if err != nil {
		if !docker.verbose {
			fmt.Println(docker.buffer.String())
		}
		return errors.New("image didn't built successfully")
	}

	return nil
}

func (docker *Docker) CreateContainer(name, image, buildDir, repoDir, tarball string) error {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{},
	}
	config := &container.Config{
		Image: image,
	}

	// archives
	hostArchivesDir := fmt.Sprintf("/tmp/%s", image)
	mountArchives := mount.Mount{
		Type:   mount.TypeBind,
		Source: hostArchivesDir,
		Target: ContainerArchivesDir,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mountArchives)

	// source
	hostSourceDir := os.Getenv("PWD")
	mountSource := mount.Mount{
		Type:   mount.TypeBind,
		Source: hostSourceDir,
		Target: ContainerSourceDir,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mountSource)

	// build
	hostBuildDir := fmt.Sprintf("%s/../%s", hostSourceDir, buildDir)
	mountBuild := mount.Mount{
		Type:   mount.TypeBind,
		Source: hostBuildDir,
		Target: ContainerBuildDir,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mountBuild)

	// repo
	if repoDir != "" {
		hostRepoDir, err := filepath.Abs(repoDir)
		if err != nil {
			return err
		}
		mountRepo := mount.Mount{
			Type:   mount.TypeBind,
			Source: hostRepoDir,
			Target: ContainerRepoDir,
		}
		hostConfig.Mounts = append(hostConfig.Mounts, mountRepo)
	}

	// mkdir
	for _, mnt := range hostConfig.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// tarball
	srcTarball := fmt.Sprintf("%s/../%s", hostSourceDir, tarball)
	dstTarball := fmt.Sprintf("%s/%s", hostBuildDir, tarball)
	if tarball != "" {
		err := os.Rename(srcTarball, dstTarball)
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

	_, err = io.Copy(docker.writer, hijack.Reader)
	if err != nil {
		return err
	}

	hijack.Close()

	inspect, err := docker.client.ContainerExecInspect(docker.ctx, response.ID)
	if err != nil {
		return err
	}

	if inspect.ExitCode != 0 {
		if !docker.verbose {
			fmt.Println(docker.buffer.String())
		}
		return errors.New("command exited with non-zero status")
	}

	return nil
}
