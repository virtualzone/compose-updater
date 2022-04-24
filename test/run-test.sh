#!/bin/bash
WORKDIR=/tmp/compose-updater-test
FAILED_TESTS=0
SUCCESSFUL_TESTS=0
MQTT_PASS=test1234

# $1 = Test name
# $2 = String to check
# $3 = Number of times the string must appear in order to succeed
function checkLogContains() {
    if ((`grep "$2" ${WORKDIR}/test.log | wc -l` != $3)); then
        echo "Failed: $1"
        ((FAILED_TESTS=FAILED_TESTS+1))
    else
        echo "Success: $1"
        ((SUCCESSFUL_TESTS=SUCCESSFUL_TESTS+1))
    fi
}

# $1 = Test name
# $2 = String to check
# $3 = Number of times the string must appear in order to succeed
function checkMqttLogContains() {
    if ((`grep "$2" ${WORKDIR}/mqtt.log | wc -l` != $3)); then
        echo "Failed: $1"
        ((FAILED_TESTS=FAILED_TESTS+1))
    else
        echo "Success: $1"
        ((SUCCESSFUL_TESTS=SUCCESSFUL_TESTS+1))
    fi
}

function runComposeUpdateAndLog() {
    sh -c "docker exec compose-test-mqtt mosquitto_sub -h 127.0.0.1 -u compose-updater -P ${MQTT_PASS} -t 'composeupdater/#' -v" &> ${WORKDIR}/mqtt.log &
    sleep 1
    ONCE=1 PRINT_SETTINGS=1 MQTT_BROKER=tcp://host.docker.internal:1883 MQTT_USERNAME=compose-updater MQTT_PASSWORD=${MQTT_PASS} ${WORKDIR}/docker-compose-watcher &> ${WORKDIR}/test.log
    sleep 1
    kill $(ps | grep "docker exec compose-test-mqtt mosquitto_sub" | grep -v "grep" | sed -e 's/^[[:space:]]*//' | cut -d' ' -f 1)
    echo "------------ Compose Updater Log ------------"
    cat ${WORKDIR}/test.log
    echo "----------------- MQTT Log ------------------"
    cat ${WORKDIR}/mqtt.log
    echo "---------------------------------------------"
}

function prepareBin() {
    if [ ! -f "./docker-compose-watcher" ]; then
        cp ../src/* ${WORKDIR}/src
        echo "Building compose watcher..."
        cd ${WORKDIR}/src && \
            go get -d -v ./... && \
            CGO_ENABLED=0 go build -ldflags="-w -s" -o "${WORKDIR}/docker-compose-watcher" .
    else
        cp ./docker-compose-watcher ${WORKDIR}/docker-compose-watcher
    fi
}

function checkDocker() {
    DOCKER_BIN=`which docker`
    if [ -z $DOCKER_BIN ]; then
        echo "Docker binary not found"
        exit 1
    fi
    if ((`docker compose version | grep 'Docker Compose version' | wc -l` != 1)); then
        echo "Docker Compose binary not found"
        exit 1
    fi
    docker ps &> /dev/null
    if (($? != 0)); then
        echo "Docker daemon not working"
        exit 1
    fi
}

function testShouldFindNoUpdates() {
    TESTNAME="Should find no updates"
    runComposeUpdateAndLog
    checkLogContains "${TESTNAME} / check c1 found" "Checking for updates of services in ${WORKDIR}/c1/compose1.yaml" 1
    checkLogContains "${TESTNAME} / check c2 found" "Checking for updates of services in ${WORKDIR}/c2/docker-compose.yml" 1
    checkLogContains "${TESTNAME} / check service test11 found" "Processing service test11 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test12 found" "Processing service test12 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test13 found" "Processing service test13 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test21 found" "Processing service test21 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test22 found" "Processing service test22 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test23 found" "Processing service test23 (requires build: false, watched: false)" 1
    checkLogContains "${TESTNAME} / check no pulls" "Pulled new image" 0
    checkLogContains "${TESTNAME} / check no builds" "Built new image" 0
    checkLogContains "${TESTNAME} / check no service restarts in c1" "Restarting services in ${WORKDIR}/c1/compose1.yaml" 0
    checkLogContains "${TESTNAME} / check no service restarts in c2" "Restarting services in ${WORKDIR}/c2/docker-compose.yml" 0
    checkMqttLogContains "${TESTNAME} / check mqtt log contains start message" "composeupdater/update start" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains done message" "composeupdater/update done" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains no restarting message" "composeupdater/update/composition/restart/start" 0
    checkMqttLogContains "${TESTNAME} / check mqtt log contains no restart complete message" "composeupdater/update/composition/restart/done" 0
}

function testShouldFindUpdateC1() {
    TESTNAME="Should find update of watcher-test-1 / test11 (c1-test11-1)"
    docker build -q --no-cache -t watcher-test-1 ${WORKDIR}
    runComposeUpdateAndLog
    checkLogContains "${TESTNAME} / check c1 found" "Checking for updates of services in ${WORKDIR}/c1/compose1.yaml" 1
    checkLogContains "${TESTNAME} / check c2 found" "Checking for updates of services in ${WORKDIR}/c2/docker-compose.yml" 1
    checkLogContains "${TESTNAME} / check service test11 found" "Processing service test11 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test12 found" "Processing service test12 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test13 found" "Processing service test13 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test21 found" "Processing service test21 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test22 found" "Processing service test22 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test23 found" "Processing service test23 (requires build: false, watched: false)" 1
    checkLogContains "${TESTNAME} / check one pull" "Pulled new image" 1
    checkLogContains "${TESTNAME} / check no builds" "Built new image" 0
    checkLogContains "${TESTNAME} / check no service restarts in c1" "Restarting services in ${WORKDIR}/c1/compose1.yaml" 1
    checkLogContains "${TESTNAME} / check no service restarts in c2" "Restarting services in ${WORKDIR}/c2/docker-compose.yml" 0
    checkMqttLogContains "${TESTNAME} / check mqtt log contains start message" "composeupdater/update start" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains done message" "composeupdater/update done" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains pulled message" "composeupdater/update/composition/service/pulled {\"composeFile\":\"${WORKDIR}/c1/compose1.yaml\",\"services\":\[{\"name\":\"test11\",\"image\":\"watcher-test-1\"}\]}" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains restarting message" "composeupdater/update/composition/restart/start {\"composeFile\":\"${WORKDIR}/c1/compose1.yaml\"" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains restart complete message" "composeupdater/update/composition/restart/done {\"composeFile\":\"${WORKDIR}/c1/compose1.yaml\"" 1
}

function testShouldFindUpdateC2() {
    TESTNAME="Should find update of watcher-test-2 / test21 (c2-test21-1)"
    docker build -q --no-cache -t watcher-test-2 ${WORKDIR}
    runComposeUpdateAndLog
    checkLogContains "${TESTNAME} / check c1 found" "Checking for updates of services in ${WORKDIR}/c1/compose1.yaml" 1
    checkLogContains "${TESTNAME} / check c2 found" "Checking for updates of services in ${WORKDIR}/c2/docker-compose.yml" 1
    checkLogContains "${TESTNAME} / check service test11 found" "Processing service test11 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test12 found" "Processing service test12 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test13 found" "Processing service test13 (requires build: true, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test21 found" "Processing service test21 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test22 found" "Processing service test22 (requires build: false, watched: true)" 1
    checkLogContains "${TESTNAME} / check service test23 found" "Processing service test23 (requires build: false, watched: false)" 1
    checkLogContains "${TESTNAME} / check one pull" "Pulled new image" 1
    checkLogContains "${TESTNAME} / check no builds" "Built new image" 0
    checkLogContains "${TESTNAME} / check no service restarts in c1" "Restarting services in ${WORKDIR}/c1/compose1.yaml" 0
    checkLogContains "${TESTNAME} / check no service restarts in c2" "Restarting services in ${WORKDIR}/c2/docker-compose.yml" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains start message" "composeupdater/update start" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains done message" "composeupdater/update done" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains pulled message" "composeupdater/update/composition/service/pulled {\"composeFile\":\"${WORKDIR}/c2/docker-compose.yml\",\"services\":\[{\"name\":\"test21\",\"image\":\"watcher-test-2\"}\]}" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains restarting message" "composeupdater/update/composition/restart/start {\"composeFile\":\"${WORKDIR}/c2/docker-compose.yml\"" 1
    checkMqttLogContains "${TESTNAME} / check mqtt log contains restart complete message" "composeupdater/update/composition/restart/done {\"composeFile\":\"${WORKDIR}/c2/docker-compose.yml\"" 1
}

echo "Working directory: ${WORKDIR}"

echo "Preparing working environment..."
checkDocker
rm -rf ${WORKDIR}
mkdir -p ${WORKDIR} ${WORKDIR}/c1 ${WORKDIR}/c2 ${WORKDIR}/src
cp ./test.Dockerfile ${WORKDIR}/Dockerfile
cp ./mosquitto.conf ${WORKDIR}
cp ./mqpass ${WORKDIR}
cp ./mqtt.Dockerfile ${WORKDIR}
PWD=$(echo ${WORKDIR} | sed 's_/_\\/_g')
cat ./c1.yaml | sed "s/\${PWD}/${PWD}/g" > ${WORKDIR}/c1/compose1.yaml
cat ./c2.yaml | sed "s/\${PWD}/${PWD}/g" > ${WORKDIR}/c2/docker-compose.yml
prepareBin

echo "Building watcher-test..."
docker build -q --no-cache -t watcher-test-1 ${WORKDIR}
docker build -q --no-cache -t watcher-test-2 ${WORKDIR}

echo "Starting MQTT broker..."
docker build -t compose-test-mqtt -f ${WORKDIR}/mqtt.Dockerfile ${WORKDIR}
docker run --rm -p 1883:1883 --name compose-test-mqtt -d compose-test-mqtt

echo "Starting composition 1..."
docker compose -f ${WORKDIR}/c1/compose1.yaml up -d --quiet-pull

echo "Starting composition 2..."
docker compose -f ${WORKDIR}/c2/docker-compose.yml up -d --quiet-pull

echo "Running integration test..."
testShouldFindNoUpdates
testShouldFindUpdateC1
testShouldFindUpdateC2

echo "Cleaning up..."
docker stop compose-test-mqtt
docker compose -f ${WORKDIR}/c1/compose1.yaml down
docker compose -f ${WORKDIR}/c2/docker-compose.yml down
rm -rf ${WORKDIR}

echo "Successful tests: ${SUCCESSFUL_TESTS}"
echo "Failed tests:     ${FAILED_TESTS}"

if ((FAILED_TESTS != 0)); then
    exit 1
fi
exit 0