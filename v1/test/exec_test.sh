#!/bin/bash

trap ctrl_c INT

function ctrl_c() {
	exit 0
}

while true; do
	echo "here we are"
    sleep 0.5
done
