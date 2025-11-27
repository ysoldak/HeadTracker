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
	FLASH_CHANNELS_LENGTH = 3
)

const FLASH_LENGTH = FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH + FLASH_NAME_LENGTH + FLASH_CHANNELS_LENGTH

type Flash struct {
	checksum      byte
	length        byte
	gyrCalOffsets [3]int32
	name          [16]byte

	// 0 - index of first channel that contains ht data, wraps around 8 output channels
	// 1 - channels order: 0=123, 1=132, ..., 6=321
	// 2 - channels enabled flags: 0b00000XXX, X=1 if channel is enabled, X=0 if disabled
	channels [3]byte
}

func NewFlash() *Flash {
	flash := Flash{}
	flash.channels[2] = 0b00000111 // all 3 channels enabled by default
	return &flash
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
	return fd.gyrCalOffsets[0] == 0 && fd.gyrCalOffsets[1] == 0 && fd.gyrCalOffsets[2] == 0
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

	offset := FLASH_HEADER_LENGTH
	if length >= byte(offset+FLASH_GYRO_CAL_LENGTH) {
		for i := range fd.gyrCalOffsets {
			fd.gyrCalOffsets[i] = toInt32(data[offset+i*4 : offset+(i+1)*4])
		}
	}

	offset += FLASH_GYRO_CAL_LENGTH
	if length >= byte(offset+FLASH_NAME_LENGTH) {
		copy(fd.name[:], data[offset:offset+FLASH_NAME_LENGTH])
	}

	offset += FLASH_NAME_LENGTH
	if length >= byte(offset+FLASH_CHANNELS_LENGTH) {
		copy(fd.channels[:], data[offset:offset+FLASH_CHANNELS_LENGTH])
	}

	return nil
}

func (fd *Flash) Store() error {
	data := make([]byte, FLASH_LENGTH)
	data[1] = FLASH_LENGTH
	// gyro calibration
	for i := range fd.gyrCalOffsets {
		fromInt32(data[FLASH_HEADER_LENGTH+i*4:FLASH_HEADER_LENGTH+(i+1)*4], fd.gyrCalOffsets[i])
	}
	// name
	nameOffset := FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH
	copy(data[nameOffset:nameOffset+FLASH_NAME_LENGTH], fd.name[:])

	// channels
	channelsOffset := FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH + FLASH_NAME_LENGTH
	copy(data[channelsOffset:channelsOffset+FLASH_CHANNELS_LENGTH], fd.channels[:])

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
