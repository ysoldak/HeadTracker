//go:build !s140v7
// +build !s140v7

package main

var channels = [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500}

func paraSetup() {
}

func paraSet(idx byte, value uint16) {
	channels[idx] = value
}

func paraSend() {
	println(channels[0], " ", channels[1], " ", channels[2])
}
