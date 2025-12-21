package main

import (
	"errors"
	"machine"
)

var errFlashWrongChecksum = errors.New("wrong checksum reading data from flash")

type Flash struct {
	gyrCalOffsets [3]int32
}

func (fd *Flash) IsEmpty() bool {
	return fd.gyrCalOffsets[0] == 0 && fd.gyrCalOffsets[1] == 0 && fd.gyrCalOffsets[2] == 0
}

func (fd *Flash) Load() error {

	data := make([]byte, 3*4+1) // each calibration parameter is int32, plus checksum
	_, err := machine.Flash.ReadAt(data, 0)
	if err != nil {
		return err
	}

	// xor all bytes, including the last checksum byte
	checksum := byte(0)
	for _, b := range data {
		checksum ^= b
	}
	if checksum != 0 {
		return errFlashWrongChecksum
	}

	fd.gyrCalOffsets[0] = toInt32(data[0:4])
	fd.gyrCalOffsets[1] = toInt32(data[4:8])
	fd.gyrCalOffsets[2] = toInt32(data[8:12])

	return nil
}

func (fd *Flash) Store() error {
	data := make([]byte, 3*4+1) // each calibration parameter is int32, plus checksum
	fromInt32(data[0:4], fd.gyrCalOffsets[0])
	fromInt32(data[4:8], fd.gyrCalOffsets[1])
	fromInt32(data[8:12], fd.gyrCalOffsets[2])
	// xor all bytes, including the last (it is zero anyway)
	checksum := byte(0)
	for _, b := range data {
		checksum ^= b
	}
	data[12] = checksum

	err := machine.Flash.EraseBlocks(0, 1)
	if err != nil {
		return err
	}
	_, err = machine.Flash.WriteAt(data, 0)
	if err != nil {
		return err
	}

	return nil
}

func toInt32(b []byte) int32 {
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func fromInt32(b []byte, v int32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
