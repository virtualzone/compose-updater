package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// DockerContainer is a docker container
type DockerContainer struct {
	ID                 string
	Name               string
	ComposeServiceName string
	ComposeFile        string
	Image              *DockerImage
}

// The getWatchedRunningContainers function returns a list of watched docker container
// including some details about the containers and their images.
func GetWatchedRunningContainers() []*DockerContainer {
	containers := []*DockerContainer{}
	containerIDs := getWatchedRunningContainerIDs()
	for _, containerID := range containerIDs {
		containers = append(containers, GetRunningContainerDetails(containerID))
	}
	return containers
}

// The getWatchedRunningContainerIDs function returns a list of IDs of running
// Docker containers labeled with docker-compose-watcher.watch=1.
func getWatchedRunningContainerIDs() []string {
	containerIDs := []string{}
	out, err := exec.Command("docker", "ps", "-a", "-q", "--no-trunc", "--filter", "label=docker-compose-watcher.watch=1").Output()
	if err != nil {
		log.Println("Failed in getWatchedRunningContainerIDs()")
		log.Fatal(err)
	}
	lines := strings.Split(string(out), "\n")
	for _, containerID := range lines {
		if strings.TrimSpace(containerID) != "" {
			containerIDs = append(containerIDs, containerID)
		}
	}
	return containerIDs
}

func GetRunningContainerDetails(id string) *DockerContainer {
	details := getRunningContainerRawDetails(id)
	return &DockerContainer{
		ID:                 id,
		Name:               getRunningContainerImageName(details),
		ComposeServiceName: getRunningContainerComposeServiceName(details),
		ComposeFile:        getRunningContainerComposeFile(details),
		Image: &DockerImage{
			ID:   getRunningContainerImageID(details),
			Hash: getRunningContainerImageHash(details),
		},
	}
}

// The getRunningContainerRawDetails returns an ordered array of details about
// a running Docker container.
func getRunningContainerRawDetails(id string) []string {
	details := []string{
		"{{.Image}}",
		"{{.Config.Image}}",
		"{{.Name}}",
		"{{index .Config.Labels \"com.docker.compose.service\"}}",
		"{{index .Config.Labels \"docker-compose-watcher.file\"}}",
		"{{index .Config.Labels \"docker-compose-watcher.dir\"}}",
	}
	formatting := strings.Join(details, "|")
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", formatting, id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerDetails('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	res := strings.Split(string(out), "|")
	for i, s := range res {
		s = strings.TrimSpace(s)
		if s == "<no value>" {
			s = ""
		}
		res[i] = s
	}
	return res
}

func getRunningContainerImageHash(rawDetails []string) string {
	return rawDetails[0]
}

func getRunningContainerImageID(rawDetails []string) string {
	return rawDetails[1]
}

func getRunningContainerImageName(rawDetails []string) string {
	return rawDetails[2]
}

func getRunningContainerComposeServiceName(rawDetails []string) string {
	return rawDetails[3]
}

func getRunningContainerComposeFile(rawDetails []string) string {
	fileName := rawDetails[4]
	if fileName == "" {
		fileName = strings.TrimSuffix(rawDetails[5], "/")
		if _, err := os.Stat(fileName + "/docker-compose.yml"); err == nil {
			fileName = fileName + "/docker-compose.yml"
		} else {
			fileName = fileName + "/docker-compose.yaml"
		}
	}
	return fileName
}
