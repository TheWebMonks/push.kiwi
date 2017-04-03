#!/usr/bin/env bash

function execute {
    curl --upload-file "$1" http://localhost:8080/$2
}

execute dev_env.sh dev_env.sh
