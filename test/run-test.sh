#!/bin/bash
WORKDIR=/tmp/compose-updater-test
FAILED_TESTS=0
SUCCESSFUL_TESTS=0

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

function runComposeUpdateAndLog() {
    ONCE=1 ${WORKDIR}/docker-compose-watcher &> ${WORKDIR}/test.log
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

echo "Working directory: ${WORKDIR}"

echo "Preparing working environment..."
checkDocker
rm -rf ${WORKDIR}
mkdir -p ${WORKDIR} ${WORKDIR}/c1 ${WORKDIR}/c2 ${WORKDIR}/src
cp ./test.Dockerfile ${WORKDIR}/Dockerfile
cp ./c1.yaml ${WORKDIR}/c1/compose1.yaml
cp ./c2.yaml ${WORKDIR}/c2/docker-compose.yml
prepareBin

echo "Building watcher-test..."
docker build -q --no-cache -t watcher-test-1 ${WORKDIR}
docker build -q --no-cache -t watcher-test-2 ${WORKDIR}

echo "Starting composition 1..."
PWD=${WORKDIR} docker compose -f ${WORKDIR}/c1/compose1.yaml up -d --quiet-pull

echo "Starting composition 2..."
PWD=${WORKDIR} docker compose -f ${WORKDIR}/c2/docker-compose.yml up -d --quiet-pull

echo "Running integration test..."

TESTNAME="Should find no updates"
runComposeUpdateAndLog
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c1/compose1.yaml" 1
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c2/docker-compose.yml" 1
checkLogContains "${TESTNAME}" "Restarted service" 0

TESTNAME="Should find update of watcher-test-1 / test1 (c1-test1-1)"
docker build -q --no-cache -t watcher-test-1 ${WORKDIR}
runComposeUpdateAndLog
checkLogContains "${TESTNAME}" "Restarted service test1 in ${WORKDIR}/c1/compose1.yaml" 1
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c1/compose1.yaml" 1
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c2/docker-compose.yml" 1
checkLogContains "${TESTNAME}" "Restarted service" 1

TESTNAME="Should find update of watcher-test-2 / test2 (c2-test1-1)"
docker build -q --no-cache -t watcher-test-2 ${WORKDIR}
runComposeUpdateAndLog
checkLogContains "${TESTNAME}" "Restarted service test1 in ${WORKDIR}/c2/docker-compose.yml" 1
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c1/compose1.yaml" 1
checkLogContains "${TESTNAME}" "Skipping Restart ${WORKDIR}/c2/docker-compose.yml" 1
checkLogContains "${TESTNAME}" "Restarted service" 1

echo "Cleaning up..."
PWD=${WORKDIR} docker compose -f ${WORKDIR}/c1/compose1.yaml down
PWD=${WORKDIR} docker compose -f ${WORKDIR}/c2/docker-compose.yml down
rm -rf ${WORKDIR}

echo "Successful tests: ${SUCCESSFUL_TESTS}"
echo "Failed tests:     ${FAILED_TESTS}"

if ((FAILED_TESTS != 0)); then
    exit 1
fi
exit 0