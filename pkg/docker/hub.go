package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Tag struct represents single JSON object received from.
//
// DockerHub API after querying it for list of tags for particular repository.
type Tag struct {
	Layer string
	Name  string
}

// GetTags function queries DockerHub API for a list of all
// available tags of a given repository.
func GetTags(repo string) ([]Tag, error) {
	tags := &[]Tag{}
	url := fmt.Sprintf("https://registry.hub.docker.com/v1/repositories/%s/tags", repo)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = response.Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, tags)
	if err != nil {
		return nil, err
	}

	return *tags, nil
}
