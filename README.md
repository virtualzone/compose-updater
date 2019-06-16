# Docker Compose Watcher

A solution for watching your Docker containers running via Docker Compose and automatically restarting the compositions whenever a base image is refreshed.

## Overview
Docker Compose Watcher is an application which continuously monitors your running docker containers. When a base image is changed, the updated version gets pulled (or built via --pull) from the registry and the docker compose composition gets restarted (via down and up -d).

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

```docker-compose-watcher.watch=1``` exposes the service to Docker Compose Watcher.

```docker-compose-watcher.dir``` specifies the path to the directory where this docker-compose.yml lives. If the file is not named docker-compose.yml, you can instead use the label ```docker-compose-watcher.file``` to specify the correct path and file name. This is necessary because it's not possible to find the docker-compose.yml from a running container.

### 2. Run Docker Compose Watcher
Run Docker Compose Watcher using compose:

```
version: '3'
services:
  watcher:
    image: virtualzone/docker-compose-watcher
    restart: always
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    environment:
      INTERVAL: 60
```

Note: You'll only need one watcher instance for all your compose services (not one per docker-compose.yml).

## Settings
Configure Watcher via environment variables (recommended) or command line arguments:

Env | Param | Default | Meaning
--- | --- | --- | ---
INTERVAL | -interval | 60 | Minutes between checks
CLEANUP | -cleanup | 1 | Run docker system prune -a -f after each run
ONCE | -once | 0 | Run once and exit
PRINT_SETTINGS | printSettings | 1 | Print settings on start

# License
GNU General Public License v3.0