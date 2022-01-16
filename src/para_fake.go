//go:build !s140v7
// +build !s140v7

package main

import "time"

var channels = [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500}

func paraSetup() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			println(time.Now().Unix(), ": ", " [", channels[0], ",", channels[1], ",", channels[2], "]")
		}
	}()
}

func paraSet(idx byte, value uint16) {
	channels[idx] = value
}

func paraSend() {
}
