#!/usr/bin/env bash

if [ "$#" -ne 3 ]; then
    echo "usage: $0 /path/to/Path of Exile/ <tree_version> <game_version>"
    echo ""
    echo "example: $0 \"/home/vilsol/.local/share/Steam/steamapps/common/Path of Exile/\" 3.21.0 3.21"
    exit 1
fi

mkdir -p "data/$3"

curl -L https://github.com/poe-tool-dev/dat-schema/releases/download/latest/schema.min.json > "data/$3/schema.min.json"

go run . "$1" "$2" "$3"

go generate -x ./...