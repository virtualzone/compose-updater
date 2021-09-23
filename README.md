# Compose Updater
A solution for watching your Docker® containers running via Docker Compose for image updates and automatically restarting the compositions whenever an image is refreshed.

## Overview
Compose Updater is an application which continuously monitors your running docker containers. When a base image is changed, the updated version gets pulled (or built via --pull) from the registry and the docker compose composition gets restarted (via down and up -d).

## Usage
### 1. Prepare your services
You'll need to add two labels to the services you want to watch:

```
version: '3'
services:
  web:
    image: nginx:alpine
    labels:
      - "docker-compose-watcher.watch=1"
      - "docker-compose-watcher.dir=/home/docker/dir"
```

```docker-compose-watcher.watch=1``` exposes the service to Compose Updater.

```docker-compose-watcher.dir``` specifies the path to the directory where this docker-compose.yml lives. If the file is not named docker-compose.yml, you can instead use the label ```docker-compose-watcher.file``` to specify the correct path and file name. This is necessary because it's not possible to find the docker-compose.yml from a running container.

### 2. Run Compose Updater
Run Docker Compose Watcher using compose:

```
version: '3'
services:
  watcher:
    image: virtualzone/compose-updater
    restart: always
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "/home/docker:/home/docker:ro"
    environment:
      INTERVAL: 60
```

It's important to mount ```/var/run/docker.sock``` and the directory your compose files reside in (```/home/docker``` in the example above).

Note: You'll only need one Compose Updater instance for all your compose services (not one per docker-compose.yml).

> :warning: **Do not include Compose Updater in same docker-compose.yml which should be updated**: Compose Updater is performing a combined restart with a complete stop of all services in the given docker-compose.yml. If Compose Updater is include, the first updated image will iniate a stop without a chance to perform start.

## Settings
Configure Compose Updater via environment variables (recommended) or command line arguments:

Env | Param | Default | Meaning
--- | --- | --- | ---
INTERVAL | -interval | 60 | Minutes between checks
CLEANUP | -cleanup | 1 | Run docker system prune -a -f after each run
ONCE | -once | 0 | Run once and exit
PRINT_SETTINGS | -printSettings | 1 | Print settings on start
UPDATE_LOG | -updateLog | '' | Log file for updates and restarts
COMPLETE_STOP | -completeStop | 0 | Restart all services in docker-compose.yml (even unmanaged)

# License
GNU General Public License v3.0

Docker® is a trademark of Docker, Inc.

This project is not affiliated with Docker, Inc.
