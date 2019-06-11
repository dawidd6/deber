package docker

import (
	"archive/tar"
	"bytes"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"os"
	"strings"
	"time"
)

const (
	// ImageMaxAge constant defines how old image can be
	//
	// If image was created ImageMaxAge time ago, then
	// it should be rebuilt
	ImageMaxAge = time.Hour * 24 * 14
)

// ImageBuildArgs struct represents arguments
// passed to ImageBuild().
type ImageBuildArgs struct {
	// Full parent image name,
	// placed in Dockerfile's FROM directive
	//
	// Example: ubuntu:bionic
	From string
	// Full to-be-built image name
	//
	// Example: deber:unstable
	Name string
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

// ImageList returns a list of images that match passed criteria.
func (docker *Docker) ImageList(prefix string) ([]string, error) {
	images := make([]string, 0)
	options := types.ImageListOptions{
		All: true,
	}

	list, err := docker.client.ImageList(docker.ctx, options)
	if err != nil {
		return nil, err
	}

	for _, v := range list {
		for _, name := range v.RepoTags {
			name = strings.TrimPrefix(name, "/")

			if strings.HasPrefix(name, prefix) {
				images = append(images, name)
			}
		}
	}

	return images, nil
}