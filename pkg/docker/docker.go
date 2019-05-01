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
	if diff > MaxImageAge {
		return true, nil
	}

	return false, nil
}

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

func (docker *Docker) ImageBuild(args BuildImageArgs) error {
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

func (docker *Docker) ContainerStart(name string) error {
	options := types.ContainerStartOptions{}

	return docker.client.ContainerStart(docker.ctx, name, options)
}

func (docker *Docker) ContainerStop(name string) error {
	timeout := ContainerStopTimeout

	return docker.client.ContainerStop(docker.ctx, name, &timeout)
}

func (docker *Docker) ContainerRemove(name string) error {
	options := types.ContainerRemoveOptions{}

	return docker.client.ContainerRemove(docker.ctx, name, options)
}

func (docker *Docker) ContainerExec(args ContainerExecArgs) error {
	cmd := []string{"bash"}
	if args.Cmd != "" {
		cmd = append(cmd, "-c", args.Cmd)
	}

	config := types.ExecConfig{
		Cmd:          cmd,
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

func (docker *Docker) ContainerDisableNetwork(name string) error {
	return docker.client.NetworkDisconnect(docker.ctx, "bridge", name, false)
}

func (docker *Docker) ContainerEnableNetwork(name string) error {
	return docker.client.NetworkConnect(docker.ctx, "bridge", name, nil)
}
