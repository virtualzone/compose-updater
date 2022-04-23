package main

import (
	"log"
	"os/exec"
	"strings"
)

// DockerImage is a docker image
type DockerImage struct {
	ID   string
	Hash string
}

func CreateDockerImageInstance(ID string) *DockerImage {
	res := &DockerImage{
		ID: ID,
	}
	res.ReadImageHash()
	return res
}

// The getImageHash function returns the hash an image.
func (i *DockerImage) ReadImageHash() error {
	out, err := exec.Command("docker", "inspect", "--type", "image", "--format", "{{.Id}}", i.ID).Output()
	if err != nil {
		log.Printf("Failed in getImageHash('%s')\n", i.ID)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
		return err
	}
	i.Hash = strings.TrimSpace(string(out))
	return nil
}

func (i *DockerImage) ExistsNewerImageHash() bool {
	newImage := CreateDockerImageInstance(i.ID)
	return i.Hash != newImage.Hash
}
