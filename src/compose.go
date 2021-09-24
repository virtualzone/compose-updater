package main

import (
	"log"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

// DockerImage is a docker image
type DockerImage struct {
	id   string
	hash string
}

// DockerContainer is a docker container
type DockerContainer struct {
	id                 string
	name               string
	composeServiceName string
	composeFile        string
	image              DockerImage
}

// ComposeYamlService is a service in a compose YAML file
type ComposeYamlService struct {
	Image string
	Build map[string]string
}

// ComposeYaml is a compose YAML file
type ComposeYaml struct {
	Services map[string]ComposeYamlService
}

// ComposeMap is a key-value map of compose file path (string) and a list of docker containers
type ComposeMap map[string]DockerContainerList

// DockerContainerList is a list of docker containers
type DockerContainerList []DockerContainer

func getWatchedComposeFiles() ComposeMap {
	files := make(map[string]DockerContainerList)
	containers := getWatchedRunningContainers()
	for _, container := range containers {
		files[container.composeFile] = append(files[container.composeFile], container)
	}
	return files
}

func getWatchedRunningContainers() []DockerContainer {
	containers := []DockerContainer{}
	containerIDs := getWatchedRunningContainerIDs()
	for _, containerID := range containerIDs {
		containers = append(containers, getRunningContainerDetails(containerID))
	}
	return containers
}

func getWatchedRunningContainerIDs() []string {
	containerIDs := []string{}
	out, err := exec.Command("docker", "ps", "-a", "-q", "--filter", "label=docker-compose-watcher.watch=1").Output()
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

func getRunningContainerDetails(id string) DockerContainer {
	return DockerContainer{
		id:                 id,
		name:               getRunningContainerImageName(id),
		composeServiceName: getRunningContainerComposeServiceName(id),
		composeFile:        getRunningContainerComposeFile(id),
		image: DockerImage{
			id:   getRunningContainerImageID(id),
			hash: getRunningContainerImageHash(id),
		},
	}
}

func getImageHash(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "image", "--format", "{{.Id}}", id).Output()
	if err != nil {
		log.Printf("Failed in getImageHash('%s')\n", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageHash(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Image}}", id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerImageHash('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageID(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Config.Image}}", id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerImageID('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerImageName(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{.Name}}", id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerImageName('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerComposeServiceName(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{index .Config.Labels \"com.docker.compose.service\"}}", id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerComposeServiceName('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func getRunningContainerComposeFile(id string) string {
	out, err := exec.Command("docker", "inspect", "--type", "container", "--format", "{{index .Config.Labels \"docker-compose-watcher.file\"}}", id).Output()
	if err != nil {
		log.Printf("Failed in getRunningContainerComposeFile('%s')", id)
		log.Printf("Result: %s\n", out)
		log.Fatal(err)
	}
	fileName := strings.TrimSpace(string(out))
	if fileName == "" {
		out, err = exec.Command("docker", "inspect", "--type", "container", "--format", "{{index .Config.Labels \"docker-compose-watcher.dir\"}}", id).Output()
		if err != nil {
			log.Printf("Failed in getRunningContainerComposeFile('%s')", id)
			log.Printf("Result: %s\n", out)
			log.Fatal(err)
		}
		fileName = strings.TrimSpace(string(out))
		if strings.Index(fileName, "/") != len(fileName)-1 {
			fileName += "/"
		}
		fileName += "docker-compose.yml"
	}
	return fileName
}

func composePull(composeFile string, serviceName string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "pull", serviceName).Run()
	if err != nil {
		return false
	}
	return true
}

func composeBuild(composeFile string, serviceName string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "build", "--pull", serviceName).Run()
	if err != nil {
		return false
	}
	return true
}

func downDockerCompose(composeFile string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "down", "--remove-orphans").Run()
	if err != nil {
		return false
	}
	return true
}

func upDockerCompose(composeFile string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "up", "-d").Run()
	if err != nil {
		return false
	}
	return true
}

func upDockerService(composeFile string, service string) bool {
	err := exec.Command("docker-compose", "-f", composeFile, "up", "-d", service).Run()
	if err != nil {
		return false
	}
	return true
}

func cleanUp() bool {
	err := exec.Command("docker", "system", "prune", "-a", "-f").Run()
	if err != nil {
		return false
	}
	return true
}

func parseComposeYaml(composeFile string) ComposeYaml {
	result := ComposeYaml{}
	data, err := exec.Command("docker-compose", "-f", composeFile, "config").Output()
	if err == nil {
		err = yaml.Unmarshal(data, &result)
	}
	return result
}
