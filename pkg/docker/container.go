package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/term"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	// ContainerStopTimeout constant represents how long Docker Engine
	// will wait for container before stopping it
	ContainerStopTimeout = time.Millisecond * 10

	// ContainerStateRunning constants defines that container is running
	ContainerStateRunning = "running"
	// ContainerStateCreated constants defines that container is created
	ContainerStateCreated = "created"
	// ContainerStateExited constants defines that container has exited
	ContainerStateExited = "exited"
	// ContainerStateRestarting constants defines that container is restarting
	ContainerStateRestarting = "restarting"
	// ContainerStatePaused constants defines that container is paused
	ContainerStatePaused = "paused"
	// ContainerStateDead constants defines that container is dead
	ContainerStateDead = "dead"
)

// ContainerCreateArgs struct represents arguments
// passed to ContainerCreate().
type ContainerCreateArgs struct {
	Mounts []mount.Mount
	Image  string
	Name   string
	User   string
}

// ContainerExecArgs struct represents arguments
// passed to ContainerExec().
type ContainerExecArgs struct {
	Interactive bool
	Name        string
	Cmd         string
	WorkDir     string
	AsRoot      bool
	Skip        bool
	Network     bool
}

// IsContainerCreated function checks if container is created
// or simply just exists.
func (docker *Docker) IsContainerCreated(name string) (bool, error) {
	list, err := docker.cli.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
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
	list, err := docker.cli.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
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
	list, err := docker.cli.ContainerList(docker.ctx, types.ContainerListOptions{All: true})
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

// ContainerCreate function creates container.
//
// It's up to the caller to make to-be-mounted directories on host.
func (docker *Docker) ContainerCreate(args ContainerCreateArgs) error {
	hostConfig := &container.HostConfig{
		Mounts: args.Mounts,
	}
	config := &container.Config{
		Image: args.Image,
		User:  args.User,
	}

	_, err := docker.cli.ContainerCreate(docker.ctx, config, hostConfig, nil, args.Name)
	if err != nil {
		return err
	}

	return nil
}

// ContainerStart function starts container, just that.
func (docker *Docker) ContainerStart(name string) error {
	options := types.ContainerStartOptions{}

	return docker.cli.ContainerStart(docker.ctx, name, options)
}

// ContainerStop function stops container, just that.
//
// It utilizes ContainerStopTimeout constant.
func (docker *Docker) ContainerStop(name string) error {
	timeout := ContainerStopTimeout

	return docker.cli.ContainerStop(docker.ctx, name, &timeout)
}

// ContainerRemove function removes container, just that.
func (docker *Docker) ContainerRemove(name string) error {
	options := types.ContainerRemoveOptions{}

	return docker.cli.ContainerRemove(docker.ctx, name, options)
}

// ContainerMounts returns mounts of created container.
func (docker *Docker) ContainerMounts(name string) ([]mount.Mount, error) {
	inspect, err := docker.cli.ContainerInspect(docker.ctx, name)
	if err != nil {
		return nil, err
	}

	mounts := make([]mount.Mount, 0)

	for _, v := range inspect.Mounts {
		mnt := mount.Mount{
			Source:   v.Source,
			Target:   v.Destination,
			Type:     v.Type,
			ReadOnly: !v.RW,
		}
		mounts = append(mounts, mnt)
	}

	return mounts, nil
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

	if args.Skip {
		return nil
	}

	if args.AsRoot {
		config.User = "root"
	}

	if args.Cmd != "" {
		config.Cmd = append(config.Cmd, "-c", args.Cmd)
	}

	err := docker.ContainerNetwork(args.Name, args.Network)
	if err != nil {
		return err
	}

	response, err := docker.cli.ContainerExecCreate(docker.ctx, args.Name, config)
	if err != nil {
		return err
	}

	hijack, err := docker.cli.ContainerExecAttach(docker.ctx, response.ID, check)
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

			err = docker.ContainerExecResize(response.ID, fd)
			if err != nil {
				return err
			}

			go docker.resizeIfChanged(response.ID, fd)
			go io.Copy(hijack.Conn, os.Stdin)
		}
	}

	io.Copy(os.Stdout, hijack.Conn)
	hijack.Close()

	if !args.Interactive {
		inspect, err := docker.cli.ContainerExecInspect(docker.ctx, response.ID)
		if err != nil {
			return err
		}

		if inspect.ExitCode != 0 {
			return errors.New("command exited with non-zero status")
		}
	}

	return nil
}

func (docker *Docker) resizeIfChanged(execID string, fd uintptr) {
	channel := make(chan os.Signal)
	signal.Notify(channel, syscall.SIGWINCH)

	for {
		<-channel
		docker.ContainerExecResize(execID, fd)
	}
}

// ContainerExecResize function resizes TTY for exec process.
func (docker *Docker) ContainerExecResize(execID string, fd uintptr) error {
	winSize, err := term.GetWinsize(fd)
	if err != nil {
		return err
	}

	options := types.ResizeOptions{
		Height: uint(winSize.Height),
		Width:  uint(winSize.Width),
	}

	err = docker.cli.ContainerExecResize(docker.ctx, execID, options)
	if err != nil {
		return err
	}

	return nil
}

// ContainerNetwork checks if container is connected to network
// and then connects it or disconnects per caller request.
func (docker *Docker) ContainerNetwork(name string, wantConnected bool) error {
	network := "bridge"
	gotConnected := false

	inspect, err := docker.cli.ContainerInspect(docker.ctx, name)
	if err != nil {
		return err
	}

	for net := range inspect.NetworkSettings.Networks {
		if net == network {
			gotConnected = true
		}
	}

	if wantConnected && !gotConnected {
		return docker.cli.NetworkConnect(docker.ctx, network, name, nil)
	}

	if !wantConnected && gotConnected {
		return docker.cli.NetworkDisconnect(docker.ctx, network, name, false)
	}

	return nil
}

// ContainerList returns a list of containers that match passed criteria.
func (docker *Docker) ContainerList(prefix string) ([]string, error) {
	containers := make([]string, 0)
	options := types.ContainerListOptions{
		All: true,
	}

	list, err := docker.cli.ContainerList(docker.ctx, options)
	if err != nil {
		return nil, err
	}

	for _, v := range list {
		for _, name := range v.Names {
			name = strings.TrimPrefix(name, "/")

			if strings.HasPrefix(name, prefix) {
				containers = append(containers, name)
			}
		}
	}

	return containers, nil
}
