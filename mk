#!/bin/bash

for path in "$@"
do
    if [[ "$path" == */ ]]
    then
        mkdir -p "$path"
    else
        DIR="$(dirname "$path")"
        mkdir -p "$DIR" && echo "" >> "$path"
    fi
done
