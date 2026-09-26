#!/usr/bin/env bash

if [ "$#" -ne 2 ]; then
    echo "usage: $0 </path/to/Path of Exile/|cdn:<patch>> <game_version>"
    echo ""
    echo "example: $0 \"/home/vilsol/.local/share/Steam/steamapps/common/Path of Exile/\" 3.21"
    echo "example: $0 cdn:3.29.3.3 3.29"
    exit 1
fi

set -ex

mkdir -p "data/$2"

curl -L https://github.com/poe-tool-dev/dat-schema/releases/download/latest/schema.min.json > "data/$2/schema.min.json"

go run . "$1" "$2"

go generate -x ./...
