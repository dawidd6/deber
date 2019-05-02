package walking

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

var StepBuild = &stepping.Step{
	Name: "build",
	Run:  Build,
	Description: []string{
		"Builds image for deber's use.",
		"This step is skipped if an image is already built.",
		"Image's parent name is derived from Debian's changelog, for example",
		"if in `debian/changelog` target distribution is `bionic`, then",
		"deber will use `ubuntu:bionic` image as a parent from Docker Hub.",
		"Image's repository name is determined by querying Docker Hub API.",
		"So, if one wants to build for other distribution than specified in",
		"`debian/changelog`, just change target distribution to whatever",
		"one desires and deber will follow.",
		"Also if image is older than 14 days, deber will try to rebuild it.",
	},
}

func Build(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(name.Image)
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt {
		isImageOld, err := dock.IsImageOld(name.Image)
		if err != nil {
			return log.FailE(err)
		}
		if !isImageOld {
			return log.SkipE()
		}
	}

	for _, repo := range []string{"debian", "ubuntu"} {
		tags, err := docker.GetTags(repo)
		if err != nil {
			return log.FailE(err)
		}

		for _, tag := range tags {
			if tag.Name == deb.TargetDist {
				from := fmt.Sprintf("%s:%s", repo, deb.TargetDist)

				log.Drop()

				args := docker.BuildImageArgs{
					From: from,
					Name: name.Image,
				}
				err := dock.ImageBuild(args)
				if err != nil {
					return log.FailE(err)
				}

				return log.DoneE()
			}
		}
	}

	return log.FailE(errors.New("distribution image not found"))
}
