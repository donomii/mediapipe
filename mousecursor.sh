#!/bin/sh
sh build_desktop_examples.sh -b
go build mosuecursor/movemouse.go
./hand_tracking_cpu --calculator_graph_config_file=mediapipe/graphs/hand_tracking/hand_tracking_desktop_live.pbtxt | ./movemouse
