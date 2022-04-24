package main

import (
	"encoding/json"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type UpdateEvent struct {
	MqttClient mqtt.Client
}

type MqttComposeServiceEvent struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type MqttComposeFileEvent struct {
	ComposeFile string                     `json:"composeFile"`
	Services    []*MqttComposeServiceEvent `json:"services"`
}

var EventBus *UpdateEvent = nil

func CreateEventBus() {
	e := &UpdateEvent{}
	e.connectMqtt()
	EventBus = e
}

func (e *UpdateEvent) OnPerformUpdatesStart() {
	log.Println("Gathering details about running containers...")
	e.publishMqtt("update", "start")
}

func (e *UpdateEvent) OnPerformUpdatesComplete() {
	log.Println("Done.")
	e.publishMqttWait("update", "done")
}

func (e *UpdateEvent) OnProcessComposeFileStart(composeFile *ComposeFile) {
	log.Printf("Checking for updates of services in %s...\n", composeFile.YamlFilePath)
	e.publishMqtt("update/composition/start", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnSkipRestartComposeFileDryMode(composeFile *ComposeFile) {
	log.Printf("Dry-Mode enabled, not restarting services in %s\n", composeFile.YamlFilePath)
	e.publishMqtt("update/composition/restart/dry", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnSkipRestartComposeFileNoUpdates(composeFile *ComposeFile) {
	log.Printf("No need to restart services in %s\n", composeFile.YamlFilePath)
	e.publishMqtt("update/composition/restart/skip", composeFile.YamlFilePath)
}

func (e *UpdateEvent) OnRestartComposeFile(composeFile *ComposeFile) {
	log.Printf("Restarting services in %s...\n", composeFile.YamlFilePath)
	msg := &MqttComposeFileEvent{
		ComposeFile: composeFile.YamlFilePath,
		Services:    e.servicesToStringArray(composeFile),
	}
	e.publishMqttJSON("update/composition/restart/start", msg)
}

func (e *UpdateEvent) OnRestartComposeFileComplete(composeFile *ComposeFile) {
	log.Printf("Restarted services in %s\n", composeFile.YamlFilePath)
	msg := &MqttComposeFileEvent{
		ComposeFile: composeFile.YamlFilePath,
		Services:    e.servicesToStringArray(composeFile),
	}
	e.publishMqttJSON("update/composition/restart/done", msg)
}

func (e *UpdateEvent) OnProcessServiceStart(service *ComposeService) {
	log.Printf("Processing service %s (requires build: %t, watched: %t)...\n", service.Name, service.RequiresBuild(), service.IsWatched())
}

func (e *UpdateEvent) OnServiceNewImageBuilt(service *ComposeService) {
	log.Printf("Built new image for service %s\n", service.Name)
	msg := &MqttComposeFileEvent{
		ComposeFile: service.ComposeFile.YamlFilePath,
		Services:    []*MqttComposeServiceEvent{e.servicesToString(service)},
	}
	e.publishMqttJSON("update/composition/service/built", msg)
}

func (e *UpdateEvent) OnServiceNewImagePulled(service *ComposeService) {
	log.Printf("Pulled new image %s for service %s\n", service.ImageName, service.Name)
	msg := &MqttComposeFileEvent{
		ComposeFile: service.ComposeFile.YamlFilePath,
		Services:    []*MqttComposeServiceEvent{e.servicesToString(service)},
	}
	e.publishMqttJSON("update/composition/service/pulled", msg)
}

func (e *UpdateEvent) OnImagePruneStart() {
	log.Println("Removing unused images...")
}

func (e *UpdateEvent) OnMqttConnect(opts *mqtt.ClientOptions) {
	log.Printf("Connecting to MQTT Broker %s...\n", opts.Servers[0])
}

func (e *UpdateEvent) connectMqtt() {
	if GlobalSettings.MqttBroker == "" {
		return
	}
	opts := mqtt.NewClientOptions().AddBroker(GlobalSettings.MqttBroker).SetClientID(GlobalSettings.MqttClientID)
	if GlobalSettings.MqttUsername != "" {
		opts.SetUsername(GlobalSettings.MqttUsername)
	}
	if GlobalSettings.MqttPassword != "" {
		opts.SetPassword(GlobalSettings.MqttPassword)
	}
	e.OnMqttConnect(opts)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Could not connect to MQTT broker: %s\n", token.Error())
	}
	GlobalSettings.MqttTopicPrefix = strings.TrimSuffix(GlobalSettings.MqttTopicPrefix, "/")
	e.MqttClient = client
}

func (e *UpdateEvent) publishMqttInternal(topic string, s string, wait bool) {
	if e.MqttClient == nil {
		return
	}
	token := e.MqttClient.Publish(GlobalSettings.MqttTopicPrefix+"/"+topic, 0, false, s)
	if wait {
		token.Wait()
	}
}

func (e *UpdateEvent) publishMqtt(topic string, s string) {
	e.publishMqttInternal(topic, s, false)
}

func (e *UpdateEvent) publishMqttWait(topic string, s string) {
	e.publishMqttInternal(topic, s, true)
}

func (e *UpdateEvent) publishMqttJSON(topic string, v interface{}) error {
	json, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.publishMqttInternal(topic, string(json), false)
	return nil
}

func (e *UpdateEvent) servicesToString(service *ComposeService) *MqttComposeServiceEvent {
	item := &MqttComposeServiceEvent{
		Name:  service.Name,
		Image: service.ImageName,
	}
	return item
}

func (e *UpdateEvent) servicesToStringArray(composeFile *ComposeFile) []*MqttComposeServiceEvent {
	services := make([]*MqttComposeServiceEvent, 0)
	for _, service := range composeFile.Services {
		services = append(services, e.servicesToString(service))
	}
	return services
}
