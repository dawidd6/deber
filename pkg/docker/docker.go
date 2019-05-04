package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
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

// IsImageBuilt function check if image with given name is built.
func (docker *Docker) IsImageBuilt(name string) (bool, error) {
	list, err := docker.client.ImageList(docker.ctx, types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].RepoTags {
			if list[i].RepoTags[j] == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsImageOld function check if image should be rebuilt.
//
// ImageMaxAge constant is utilized here.
func (docker *Docker) IsImageOld(name string) (bool, error) {
	inspect, _, err := docker.client.ImageInspectWithRaw(docker.ctx, name)
	if err != nil {
		return false, err
	}

	created, err := time.Parse(time.RFC3339Nano, inspect.Created)
	if err != nil {
		return false, err
	}

	diff := time.Since(created)
	if diff > ImageMaxAge {
		return true, nil
	}

	return false, nil
}

// IsContainerCreated function checks if container is created
// or simply just exists.
func (docker *Docker) IsContainerCreated(name string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+name {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsContainerStarted function checks
// if container's state == ContainerStateRunning.
func (docker *Docker) IsContainerStarted(name string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+name {
				if list[i].State == ContainerStateRunning {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// IsContainerStopped function checks
// if container's state != ContainerStateRunning.
func (docker *Docker) IsContainerStopped(name string) (bool, error) {
	list, err := docker.client.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
	}

	for i := range list {
		for j := range list[i].Names {
			if list[i].Names[j] == "/"+name {
				if list[i].State == ContainerStateRunning {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

// ImageBuild function build image from dockerfile
// and prints output to Stdout.
func (docker *Docker) ImageBuild(args ImageBuildArgs) error {
	dockerfile, err := dockerfileParse(args.From)
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
		Tags:       []string{args.Name},
		Remove:     true,
		PullParent: true,
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

	_, _, err = docker.client.ImageInspectWithRaw(docker.ctx, args.Name)
	if err != nil {
		return errors.New("image didn't built successfully")
	}

	return nil
}

// ContainerCreate function creates container and
// makes required directories and host system.
func (docker *Docker) ContainerCreate(args ContainerCreateArgs) error {
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: args.SourceDir,
				Target: ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: args.BuildDir,
				Target: ContainerBuildDir,
			}, {
				Type:   mount.TypeBind,
				Source: args.CacheDir,
				Target: ContainerCacheDir,
			}, {
				Type:   mount.TypeBind,
				Source: args.ArchiveDir,
				Target: ContainerArchiveDir,
			},
		},
	}
	config := &container.Config{
		Image: args.Image,
		User:  args.User,
	}

	// mkdir
	for _, mnt := range hostConfig.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	_, err := docker.client.ContainerCreate(docker.ctx, config, hostConfig, nil, args.Name)
	if err != nil {
		return err
	}

	return nil
}

// ContainerStart function starts container, just that.
func (docker *Docker) ContainerStart(name string) error {
	options := types.ContainerStartOptions{}

	return docker.client.ContainerStart(docker.ctx, name, options)
}

// ContainerStop function stops container, just that.
//
// It utilizes ContainerStopTimeout constant.
func (docker *Docker) ContainerStop(name string) error {
	timeout := ContainerStopTimeout

	return docker.client.ContainerStop(docker.ctx, name, &timeout)
}

// ContainerRemove function removes container, just that.
func (docker *Docker) ContainerRemove(name string) error {
	options := types.ContainerRemoveOptions{}

	return docker.client.ContainerRemove(docker.ctx, name, options)
}

// ContainerExec function executes a command in running container.
//
// Command is executed in bash shell by default.
//
// Command can be executed as root.
//
// Command can be executed interactively.
//
// Command can be empty, in that case just bash is executed.
func (docker *Docker) ContainerExec(args ContainerExecArgs) error {
	config := types.ExecConfig{
		Cmd:          []string{"bash"},
		WorkingDir:   args.WorkDir,
		AttachStdin:  args.Interactive,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	check := types.ExecStartCheck{
		Tty:    true,
		Detach: false,
	}

	if args.AsRoot {
		config.User = "root"
	}

	if args.Cmd != "" {
		config.Cmd = append(config.Cmd, "-c", args.Cmd)
	}

	response, err := docker.client.ContainerExecCreate(docker.ctx, args.Name, config)
	if err != nil {
		return err
	}

	hijack, err := docker.client.ContainerExecAttach(docker.ctx, response.ID, check)
	if err != nil {
		return err
	}

	if args.Interactive {
		fd := os.Stdin.Fd()

		if term.IsTerminal(fd) {
			oldState, err := term.MakeRaw(fd)
			if err != nil {
				return err
			}
			defer term.RestoreTerminal(fd, oldState)

			args := ContainerExecResizeArgs{
				Fd:     fd,
				ExecID: response.ID,
			}
			err = docker.ContainerExecResize(args)
			if err != nil {
				return err
			}

			go docker.resizeIfChanged(args)
			go io.Copy(hijack.Conn, os.Stdin)
		}
	}

	io.Copy(os.Stdout, hijack.Conn)
	hijack.Close()

	if !args.Interactive {
		inspect, err := docker.client.ContainerExecInspect(docker.ctx, response.ID)
		if err != nil {
			return err
		}

		if inspect.ExitCode != 0 {
			return errors.New("command exited with non-zero status")
		}
	}

	return nil
}

func (docker *Docker) resizeIfChanged(args ContainerExecResizeArgs) {
	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGWINCH)

	for {
		<-channel
		docker.ContainerExecResize(args)
	}
}

// ContainerExecResize function resizes TTY for exec process.
func (docker *Docker) ContainerExecResize(args ContainerExecResizeArgs) error {
	winSize, err := term.GetWinsize(args.Fd)
	if err != nil {
		return err
	}

	options := types.ResizeOptions{
		Height: uint(winSize.Height),
		Width:  uint(winSize.Width),
	}

	err = docker.client.ContainerExecResize(docker.ctx, args.ExecID, options)
	if err != nil {
		return err
	}

	return nil
}

// ContainerDisableNetwork function disconnects "bridge" network from container.
func (docker *Docker) ContainerDisableNetwork(name string) error {
	return docker.client.NetworkDisconnect(docker.ctx, "bridge", name, false)
}

// ContainerEnableNetwork function connects "bridge" network to container.
func (docker *Docker) ContainerEnableNetwork(name string) error {
	return docker.client.NetworkConnect(docker.ctx, "bridge", name, nil)
}
