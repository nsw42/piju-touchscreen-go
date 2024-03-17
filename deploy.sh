#! /bin/bash

set -e 

ssh piju "rm -f piju-touchscreen-go"
scp dist/piju-touchscreen-go piju:
ssh piju "ps wwaux | grep piju-touchscreen-go | grep -v grep | awk '{print $2}' | xargs kill"
