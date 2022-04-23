package main

import (
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

func (s *ComposeService) RequiresBuild() bool {
	return len(s.BuildInfo) > 0
}

func (s *ComposeService) IsWatched() bool {
	return s.Instance != nil
}

// ATTENTION: docker compose restart does not use an updated image in Docker Compose V2 yet.
func (s *ComposeService) Restart() bool {
	err := exec.Command("docker", "compose", "-f", s.ComposeFile.YamlFilePath, "up", "-d", "--no-deps", s.Name).Run()
	if err != nil {
		return false
	}
	return true
}
