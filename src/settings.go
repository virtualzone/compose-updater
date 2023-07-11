package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

// Settings holds the program runtime configuration
type Settings struct {
	Cleanup         bool
	Dry             bool
	Help            bool
	Interval        int64
	Once            bool
	PrintSettings   bool
	UpdateLog       string
	Build           bool
	MqttBroker      string
	MqttClientID    string
	MqttTopicPrefix string
	MqttUsername    string
	MqttPassword    string
	//CompleteStop  bool
}

var GlobalSettings *Settings = nil

func ReadSettings() {
	s := &Settings{}
	s.boolFlagEnv(&s.Cleanup, "cleanup", "CLEANUP", false, "run docker system prune at the end")
	s.boolFlagEnv(&s.Dry, "dry", "DRY", false, "dry run: check and pull, but don't restart")
	s.boolFlagEnv(&s.Help, "help", "HELP", false, "print usage instructions")
	s.int64FlagEnv(&s.Interval, "interval", "INTERVAL", 60, "interval in minutes between runs")
	s.boolFlagEnv(&s.Once, "once", "ONCE", false, "run once and exit, do not run in background")
	s.boolFlagEnv(&s.PrintSettings, "printSettings", "PRINT_SETTINGS", false, "print used settings")
	s.stringFlagEnv(&s.UpdateLog, "updateLog", "UPDATE_LOG", "", "update log file")
	//s.boolFlagEnv(&s.CompleteStop, "completeStop", "COMPLETE_STOP", false, "Restart all services in docker-compose.yml (even unmanaged) after a new image is pulled")
	s.boolFlagEnv(&s.Build, "build", "BUILD", false, "Rebuild images of services with 'build' definition")
	s.stringFlagEnv(&s.MqttBroker, "mqttBroker", "MQTT_BROKER", "", "MQTT Broker address (i.e. tcp://127.0.0.1:1883)")
	s.stringFlagEnv(&s.MqttClientID, "mqttClientId", "MQTT_CLIENT_ID", "composeupdater", "MQTT Client ID")
	s.stringFlagEnv(&s.MqttTopicPrefix, "mqttTopicPrefix", "MQTT_TOPIC_PREFIX", "composeupdater", "MQTT Topic Prefix")
	s.stringFlagEnv(&s.MqttUsername, "mqttUsername", "MQTT_USERNAME", "", "MQTT Username")
	s.stringFlagEnv(&s.MqttPassword, "mqttPassword", "MQTT_PASSWORD", "", "MQTT Password")
	flag.Parse()
	GlobalSettings = s
}

func (settings *Settings) boolFlagEnv(p *bool, name string, env string, value bool, usage string) {
	flag.BoolVar(p, name, value, usage+" (env "+env+")")
	val := os.Getenv(env)
	if val != "" {
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatal(err)
		}
		*p = b
	}
}

func (settings *Settings) int64FlagEnv(p *int64, name string, env string, value int64, usage string) {
	flag.Int64Var(p, name, value, usage+" (env "+env+")")
	val := os.Getenv(env)
	if val != "" {
		i, err := strconv.ParseInt(val, 10, 0)
		if err != nil {
			log.Fatal(err)
		}
		*p = i
	}
}

func (settings *Settings) stringFlagEnv(p *string, name string, env string, value string, usage string) {
	flag.StringVar(p, name, value, usage+" (env "+env+")")
	val := os.Getenv(env)
	if val != "" {
		*p = val
	}
}

func (settings *Settings) Print() {
	log.Println("Using settings:")
	log.Printf("    cleanup ......... %t\n", settings.Cleanup)
	log.Printf("    dry ............. %t\n", settings.Dry)
	log.Printf("    help ............ %t\n", settings.Help)
	log.Printf("    interval ........ %d\n", settings.Interval)
	log.Printf("    once ............ %t\n", settings.Once)
	log.Printf("    printSettings ... %t\n", settings.PrintSettings)
	//log.Printf("    completeStop .... %t\n", settings.CompleteStop)
	log.Printf("    build ........... %t\n", settings.Build)
	log.Printf("    mqttBroker ...... %s\n", settings.MqttBroker)
	log.Printf("    mqttClientId .... %s\n", settings.MqttClientID)
	log.Printf("    mqttTopicPrefix . %s\n", settings.MqttTopicPrefix)
	log.Printf("    mqttUsername .... %s\n", settings.MqttUsername)
	log.Printf("    mqttPassword .... %s\n", "(hidden)")
}
