package main

import (
	"errors"
	"machine"
)

var errFlashWrongChecksum = errors.New("wrong checksum reading data from flash")

type Flash struct {
	roll, pitch, yaw int32
	name             [16]byte
}

func (fd *Flash) SetName(name string) {
	n := 0
	for n < len(name) && n < len(fd.name) {
		fd.name[n] = byte(name[n])
		n++
	}
	for n < len(fd.name) {
		fd.name[n] = 0
		n++
	}
}

func (fd *Flash) Name() string {
	n := 0
	for n < len(fd.name) && fd.name[n] != 0 {
		n++
	}
	return string(fd.name[:n])
}

func (fd *Flash) IsEmpty() bool {
	return fd.roll == 0 && fd.pitch == 0 && fd.yaw == 0
}

func (fd *Flash) Load() error {

	data := make([]byte, 3*4+16+1) // each calibration parameter is int32, plus name, plus checksum
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

	fd.roll = toInt32(data[0:4])
	fd.pitch = toInt32(data[4:8])
	fd.yaw = toInt32(data[8:12])
	copy(fd.name[:], data[12:28])

	return nil
}

func (fd *Flash) Store() error {
	data := make([]byte, 3*4+16+1) // each calibration parameter is int32, plus name, plus checksum
	fromInt32(data[0:4], fd.roll)
	fromInt32(data[4:8], fd.pitch)
	fromInt32(data[8:12], fd.yaw)
	copy(data[12:28], fd.name[:])

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
