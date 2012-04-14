#!/bin/bash

function CreateSelfRef() {
    selfsrc='src/github.com/huin/dcpu16go'
    if [ ! -e "$selfsrc" ]; then
        mkdir -p `dirname "$selfsrc"` || exit 1
        ln -s "$PWD" "$selfsrc"
        echo '[OK]'
    elif [ ! -L "$selfsrc" ]; then
        echo '[Failed]'
        echo "$selfsrc is not a symlink, this should point to the git clone."
        echo "Maybe remove it and re-run this script."
        exit 1
    else
        echo '[Already OK]'
    fi
}

function CreateEnv() {
    echo "export GOPATH='$PWD'" > env || return $?
    echo '[OK]'
    echo '`source '"$PWD"'/env` to set GOPATH.'
}

echo 'Configuring your git clone for development. This will allow you to modify the code within.'

echo -n 'Initializing src directory... '
CreateSelfRef || exit 1

echo -n 'Creating env... '
CreateEnv || exit 1
