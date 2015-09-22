#!/usr/bin/env bash

echo -ne '|                    | (0%) -> cube\r'
go build -o cube cube.go
echo -ne '|##########          | (50%) -> joystick\r'
go build -o joystick joystick.go
echo -ne '|####################| (100%) ----------> FINISHED\r'
