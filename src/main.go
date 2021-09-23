package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// Settings holds the program runtime configuration
type Settings struct {
	cleanup       bool
	dry           bool
	help          bool
	interval      int64
	once          bool
	printSettings bool
	updateLog     string
	completeStop       bool
}

var (
	UpdateLog *log.Logger
)

func initLogger(fileName string) {
	target := ioutil.Discard
	if fileName != "" {
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", err)
		}
		target = file
	}
	UpdateLog = log.New(target, "", log.Ldate|log.Ltime)
}

func (composeFiles *ComposeMap) getNumContainers() int {
	var numContainers = 0
	for _, containers := range *composeFiles {
		numContainers += len(containers)
	}
	return numContainers
}

func (composeFiles *ComposeMap) updateAllContainers() {
	for composeFile, containers := range *composeFiles {
		log.Printf("Processing compose file %s\n", composeFile)
		yaml := parseComposeYaml(composeFile)
		for _, container := range containers {
			yamlPart := yaml.Services[container.composeServiceName]
			var res bool
			if len(yamlPart.Build) > 0 {
				log.Printf("Building and pulling for service %s ... ", container.composeServiceName)
				res = composeBuild(composeFile, container.composeServiceName)
			} else {
				log.Printf("Pulling for service %s ... ", container.composeServiceName)
				res = composePull(composeFile, container.composeServiceName)
			}
			if res {
				log.Println("OK")
			} else {
				log.Println("Failed")
			}
		}
	}
}

func (containers *DockerContainerList) needsRestart() bool {
	var needsRestart = false
	for _, container := range *containers {
		var newImageHash = (container.image.hash != getImageHash(container.image.id))
		if newImageHash {
			UpdateLog.Printf("Updated service %s in %s", container.composeServiceName, container.composeFile)
			needsRestart = true
		}
	}
	return needsRestart
}

func (composeFiles *ComposeMap) checkPerformRestart(settings *Settings) {
	for composeFile, containers := range *composeFiles {
		if containers.needsRestart() {
			log.Printf("Restarting %s ... ", composeFile)
			if (*settings).completeStop {
				downDockerCompose(composeFile)
			}
			upDockerCompose(composeFile)
			UpdateLog.Printf("Restarted %s", composeFile)
			log.Println("OK")
		} else {
			log.Printf("Skipping %s\n", composeFile)
		}
	}
}

func boolFlagEnv(p *bool, name string, env string, value bool, usage string) {
	flag.BoolVar(p, name, value, usage+" (env "+env+")")
	if os.Getenv(env) != "" {
		*p = true
	}
}

func int64FlagEnv(p *int64, name string, env string, value int64, usage string) {
	flag.Int64Var(p, name, value, usage+" (env "+env+")")
	if os.Getenv(env) != "" {
		i, _ := strconv.ParseInt(os.Getenv(env), 10, 0)
		*p = i
	}
}

func stringFlagEnv(p *string, name string, env string, value string, usage string) {
	flag.StringVar(p, name, value, usage+" (env "+env+")")
	if os.Getenv(env) != "" {
		*p = os.Getenv(env)
	}
}

func getSettings() *Settings {
	settings := new(Settings)
	boolFlagEnv(&settings.cleanup, "cleanup", "CLEANUP", false, "run docker system prune at the end")
	boolFlagEnv(&settings.dry, "dry", "DRY", false, "dry run: check and pull, but don't restart")
	boolFlagEnv(&settings.help, "help", "HELP", false, "print usage instructions")
	int64FlagEnv(&settings.interval, "interval", "INTERVAL", 60, "interval in minutes between runs")
	boolFlagEnv(&settings.once, "once", "ONCE", true, "run once and exit, do not run in background")
	boolFlagEnv(&settings.printSettings, "printSettings", "PRINT_SETTINGS", false, "print used settings")
	stringFlagEnv(&settings.updateLog, "updateLog", "UPDATE_LOG", "", "update log file")
	boolFlagEnv(&settings.completeStop, "completeStop", "COMPLETE_STOP", false, "Restart all services in docker-compose.yml (even unmanaged)")
	flag.Parse()
	return settings
}

func (settings *Settings) print() {
	log.Println("Using settings:")
	log.Printf("    cleanup ......... %t\n", settings.cleanup)
	log.Printf("    dry ............. %t\n", settings.dry)
	log.Printf("    help ............ %t\n", settings.help)
	log.Printf("    interval ........ %d\n", settings.interval)
	log.Printf("    once ............ %t\n", settings.once)
	log.Printf("    printSettings ... %t\n", settings.printSettings)
	log.Printf("    updateLog ....... %s\n", settings.updateLog)
	log.Printf("    completeStop ....... %s\n", settings.completeStop)
}

func performUpdates(settings *Settings) {
	log.Println("Building docker compose overview...")
	composeFiles := getWatchedComposeFiles()
	log.Printf("Found %d compose files with %d watched containers.\n", len(composeFiles), composeFiles.getNumContainers())
	log.Println("Trying to update containers...")
	composeFiles.updateAllContainers()
	log.Println("Updating docker compose overview...")
	composeFiles = getWatchedComposeFiles()
	if !(*settings).dry {
		composeFiles.checkPerformRestart(settings)
	} else {
		log.Println("Skipping actual restart (dry run).")
	}
	if (*settings).cleanup {
		if !(*settings).dry {
			cleanUp()
		} else {
			log.Println("Skipping clean-up (dry run).")
		}
	}
	log.Println("Done.")
}

func printHeader() {
	log.Printf("Compose Updater %s\n", BuildVersion)
}

func mainLoop(settings *Settings) {
	for {
		performUpdates(settings)
		if (*settings).once {
			return
		}
		log.Printf("Waiting %d minutes until next execution...\n", (*settings).interval)
		time.Sleep(time.Duration((*settings).interval) * time.Minute)
	}
}

func main() {
	printHeader()
	var settings = getSettings()
	if (*settings).help {
		flag.Usage()
		return
	}
	if (*settings).printSettings {
		settings.print()
	}
	initLogger((*settings).updateLog)
	mainLoop(settings)
}
