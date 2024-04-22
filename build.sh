#!/use/bin/bash

set -eux

DIST_PATH=$1
GIT_TAG=$2
FILENAME=insight-api.${GIT_TAG}
OLD_PWD=$(pwd)

go build -o ${GIT_TAG}/${FILENAME} .