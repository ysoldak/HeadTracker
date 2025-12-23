package main

import (
	"errors"
	"machine"
)

var errFlashWrongChecksum = errors.New("wrong checksum reading data from flash")
var errFlashWrongLength = errors.New("unsupported flash data length")

const (
	FLASH_HEADER_LENGTH   = 2     // checksum + length
	FLASH_GYRO_CAL_LENGTH = 3 * 4 // gyro calibration offsets (int32 each)
	FLASH_LENGTH          = FLASH_HEADER_LENGTH + FLASH_GYRO_CAL_LENGTH
)

type Flash struct {
	checksum      byte
	length        byte
	gyrCalOffsets [3]int32
}

func NewFlash() *Flash {
	return &Flash{
		checksum:      FLASH_LENGTH,
		length:        FLASH_LENGTH,
		gyrCalOffsets: [3]int32{0, 0, 0},
	}
}

func (fd *Flash) IsEmpty() bool {
	return fd.gyrCalOffsets[0] == 0 && fd.gyrCalOffsets[1] == 0 && fd.gyrCalOffsets[2] == 0
}

func (fd *Flash) Load() error {

	data := make([]byte, FLASH_LENGTH)

	_, err := machine.Flash.ReadAt(data[:], 0)
	if err != nil {
		return err
	}

	// validate length
	length := data[1]
	if length == 0 || length > FLASH_LENGTH {
		return errFlashWrongLength
	}

	// xor all bytes, but the first
	checksum := byte(0)
	for _, b := range data[1:length] {
		checksum ^= b
	}
	if checksum != data[0] {
		return errFlashWrongChecksum
	}

	// read gyro calibration, best effort
	if length < FLASH_HEADER_LENGTH+FLASH_GYRO_CAL_LENGTH {
		println("Incomplete flash data, length:", length)
		return nil // this is fine, just no gyro calibration
	}
	offset := FLASH_HEADER_LENGTH
	for i := range 3 {
		fd.gyrCalOffsets[i] = toInt32(data[offset+i*4 : offset+(i+1)*4])
	}

	println("Loaded from flash:", fd.gyrCalOffsets[0], fd.gyrCalOffsets[1], fd.gyrCalOffsets[2])

	return nil
}

func (fd *Flash) Store() error {

	data := make([]byte, FLASH_LENGTH)

	data[1] = FLASH_LENGTH
	// gyro calibration
	offset := FLASH_HEADER_LENGTH
	for i := range 3 {
		fromInt32(data[offset+i*4:offset+(i+1)*4], fd.gyrCalOffsets[i])
	}
	// xor all bytes, but the first
	checksum := byte(0)
	for _, b := range data[1:] {
		checksum ^= b
	}
	data[0] = checksum

	err := machine.Flash.EraseBlocks(0, 1)
	if err != nil {
		return err
	}
	_, err = machine.Flash.WriteAt(data[:], 0)
	if err != nil {
		return err
	}

	println("Stored to flash:", fd.gyrCalOffsets[0], fd.gyrCalOffsets[1], fd.gyrCalOffsets[2])

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
