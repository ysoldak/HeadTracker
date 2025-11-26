package main

import (
	"errors"
	"machine"
)

var errFlashWrongChecksum = errors.New("wrong checksum reading data from flash")
var errFlashWrongLength = errors.New("unsupported flash data length")

const (
	FLASH_HEADER_LENGTH   = 2
	FLASH_GYRO_CAL_LENGTH = 3 * 4
	FLASH_NAME_LENGTH     = 16
	FLASH_LENGTH          = FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH + FLASH_NAME_LENGTH // checksum, length, roll, pitch, yaw, name
)

type Flash struct {
	checksum         byte
	length           byte
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
	data := make([]byte, FLASH_LENGTH)
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

	length := data[1]
	if length == 0 || length > FLASH_LENGTH {
		return errFlashWrongLength
	}

	if length >= FLASH_HEADER_LENGTH+FLASH_GYRO_CAL_LENGTH {
		offset := FLASH_HEADER_LENGTH
		fd.roll = toInt32(data[offset+0*4 : offset+1*4])
		fd.pitch = toInt32(data[offset+1*4 : offset+2*4])
		fd.yaw = toInt32(data[offset+2*4 : offset+3*4])
	}
	if length >= FLASH_HEADER_LENGTH+FLASH_GYRO_CAL_LENGTH+FLASH_NAME_LENGTH {
		offset := FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH
		copy(fd.name[:], data[offset:offset+FLASH_NAME_LENGTH])
	}

	return nil
}

func (fd *Flash) Store() error {
	data := make([]byte, FLASH_LENGTH)
	data[1] = FLASH_LENGTH
	// gyro calibration
	fromInt32(data[FLASH_HEADER_LENGTH+0*4:FLASH_HEADER_LENGTH+1*4], fd.roll)
	fromInt32(data[FLASH_HEADER_LENGTH+1*4:FLASH_HEADER_LENGTH+2*4], fd.pitch)
	fromInt32(data[FLASH_HEADER_LENGTH+2*4:FLASH_HEADER_LENGTH+3*4], fd.yaw)
	// name
	nameOffset := FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH
	copy(data[nameOffset:nameOffset+FLASH_NAME_LENGTH], fd.name[:])

	// xor all bytes, including the first (it is zero anyway)
	checksum := byte(0)
	for _, b := range data {
		checksum ^= b
	}
	data[0] = checksum

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
