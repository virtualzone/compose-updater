package main

import (
	"log"
	"os/exec"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	YamlFilePath string
	Services     map[string]*ComposeService `yaml:"services"`
}

type ComposeRuntimeInfo struct {
	ContainerID   string `yaml:"ID"`
	ContainerName string `yaml:"Name"`
	ServiceName   string `yaml:"Service"`
}

func ParseComposeYaml(yamlFilePath string) (*ComposeFile, error) {
	result := &ComposeFile{}
	data, err := exec.Command("docker", "compose", "-f", yamlFilePath, "config").Output()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	result.YamlFilePath = yamlFilePath
	return result, nil
}

func (f *ComposeFile) AttachRuntimeInfo() {
	runtimeInfos, err := f.getRuntimeInfo()
	if err != nil {
		log.Fatalf("Could not get runtime info: %s\n", err)
	}
	for _, runtimeInfo := range runtimeInfos {
		if service, ok := f.Services[runtimeInfo.ServiceName]; ok {
			service.Name = runtimeInfo.ServiceName
			service.ContainerID = runtimeInfo.ContainerID
			service.ContainerName = runtimeInfo.ContainerName
			service.ComposeFile = f
		}
	}
}

func (f *ComposeFile) getRuntimeInfo() ([]*ComposeRuntimeInfo, error) {
	var result []*ComposeRuntimeInfo
	data, err := exec.Command("docker", "compose", "-f", f.YamlFilePath, "ps", "--format", "json").Output()
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (f *ComposeFile) Down() bool {
	err := exec.Command("docker", "compose", "-f", f.YamlFilePath, "down", "--remove-orphans").Run()
	if err != nil {
		return false
	}
	return true
}

func (f *ComposeFile) Up() bool {
	err := exec.Command("docker", "compose", "-f", f.YamlFilePath, "up", "-d").Run()
	if err != nil {
		return false
	}
	return true
}
