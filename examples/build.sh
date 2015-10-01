#!/usr/bin/env bash

echo -ne '|                    | (0%) -> cube\r'
go build -o cube cube.go
echo -ne '|######              | (33%) --> joystick\r'
go build -o joystick joystick.go
echo -ne '|#############       | (66%) ---> lights_forward\r'
go build -o lights_forward lights_forward.go
echo -ne '|####################| (100%) ----------> FINISHED\r'
