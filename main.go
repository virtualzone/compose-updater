package main

import (
	"flag"
	"fmt"
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
}

func (composeFiles *ComposeMap) getNumContainers() int {
	var numContainers = 0
	for _, containers := range *composeFiles {
		numContainers += len(containers)
	}
	return numContainers
}

func (composeFiles *ComposeMap) updateAllContainers() {
	for _, containers := range *composeFiles {
		for _, container := range containers {
			fmt.Printf("Pulling image %s ... ", container.image.id)
			var res = pullImage(container.image.id)
			if res {
				fmt.Println("OK")
			} else {
				fmt.Println("Failed")
			}
		}
	}
}

func (containers *DockerContainerList) needsRestart() bool {
	var needsRestart = false
	for _, container := range *containers {
		needsRestart = needsRestart || (container.image.hash != getImageHash(container.image.id))
	}
	return needsRestart
}

func (composeFiles *ComposeMap) checkPerformRestart() {
	for composeFile, containers := range *composeFiles {
		if containers.needsRestart() {
			fmt.Printf("Restarting %s ... ", composeFile)
			downDockerCompose(composeFile)
			upDockerCompose(composeFile)
			fmt.Println("OK")
		} else {
			fmt.Printf("Skipping %s\n", composeFile)
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

func getSettings() *Settings {
	settings := new(Settings)
	boolFlagEnv(&settings.cleanup, "cleanup", "CLEANUP", false, "run docker system prune at the end")
	boolFlagEnv(&settings.dry, "dry", "DRY", false, "dry run: check and pull, but don't restart")
	boolFlagEnv(&settings.help, "help", "HELP", false, "print usage instructions")
	int64FlagEnv(&settings.interval, "interval", "INTERVAL", 60, "interval in minutes between runs")
	boolFlagEnv(&settings.once, "once", "ONCE", true, "run once and exit, do not run in background")
	boolFlagEnv(&settings.printSettings, "printSettings", "PRINT_SETTINGS", false, "print used settings")
	flag.Parse()
	return settings
}

func (settings *Settings) print() {
	fmt.Println("Using settings:")
	fmt.Printf("    cleanup ......... %t\n", settings.cleanup)
	fmt.Printf("    dry ............. %t\n", settings.dry)
	fmt.Printf("    help ............ %t\n", settings.help)
	fmt.Printf("    interval ........ %d\n", settings.interval)
	fmt.Printf("    once ............ %t\n", settings.once)
	fmt.Printf("    printSettings ... %t\n", settings.printSettings)
}

func performUpdates(settings *Settings) {
	fmt.Println("Building docker compose overview...")
	composeFiles := getWatchedComposeFiles()
	fmt.Printf("Found %d compose files with %d watched containers.\n", len(composeFiles), composeFiles.getNumContainers())
	fmt.Println("Trying to update containers...")
	composeFiles.updateAllContainers()
	fmt.Println("Updating docker compose overview...")
	composeFiles = getWatchedComposeFiles()
	if !(*settings).dry {
		composeFiles.checkPerformRestart()
	}
	if (*settings).cleanup && !(*settings).dry {
		cleanUp()
	}
	fmt.Println("Done.")
}

func printHeader() {
	fmt.Printf("Docker Compose Watcher %s\n", BuildVersion)
	fmt.Println("https://github.com/virtualzone/docker-compose-watcher")
	fmt.Println("=====================================================")
}

func mainLoop(settings *Settings) {
	for {
		performUpdates(settings)
		if (*settings).once {
			return
		}
		fmt.Printf("Waiting %d minutes until next execution...\n", (*settings).interval)
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
	mainLoop(settings)
}
