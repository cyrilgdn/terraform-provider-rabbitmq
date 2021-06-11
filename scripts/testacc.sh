#!/bin/bash
set -e

SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

source "$SCRIPT_DIR/test.env"

setup() {
    docker-compose -f "$SCRIPT_DIR/docker-compose.yml" up -d
    "$SCRIPT_DIR"/wait-rabbitmq-docker.sh
}

run() {
    go test -count=1 ./rabbitmq -v -timeout 120m

    # keep the return value for the scripts to fail and clean properly
    return $?
}

cleanup() {
    docker-compose -f "$SCRIPT_DIR/docker-compose.yml" down
}

testacc() {
    setup

    run || (cleanup && exit 1)

    cleanup
}


case "$1" in
    "setup")
        setup
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        testacc
        ;;
esac
