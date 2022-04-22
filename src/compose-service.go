package main

import (
	"log"
	"os/exec"
)

// ComposeYamlService is a service in a compose YAML file
type ComposeService struct {
	Name          string
	ContainerID   string
	ContainerName string
	ImageName     string            `yaml:"image"`
	BuildInfo     map[string]string `yaml:"build"`
	Instance      *DockerContainer
	ComposeFile   *ComposeFile
}

func (s *ComposeService) Pull() bool {
	log.Println("Running: ", "docker", "compose", "-f", s.ComposeFile.YamlFilePath, "pull", s.Name)
	err := exec.Command("docker", "compose", "-f", s.ComposeFile.YamlFilePath, "pull", s.Name).Run()
	if err != nil {
		return false
	}
	return true
}

func (s *ComposeService) Build() bool {
	err := exec.Command("docker", "compose", "-f", s.ComposeFile.YamlFilePath, "build", "--pull", s.Name).Run()
	if err != nil {
		return false
	}
	return true
}

/*
func (s *ComposeService) Restart() bool {
	log.Println("docker", "compose", "-f", s.ComposeFile.YamlFilePath, "up", "-d", "--no-deps", s.Name)
	out, err := exec.Command("docker", "compose", "-f", s.ComposeFile.YamlFilePath, "up", "-d", "--no-deps", s.Name).CombinedOutput()
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("-----> " + string(out))
	return true
}
*/
