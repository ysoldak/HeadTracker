package main

import (
	"errors"
	"machine"
)

var errFlashWrongChecksum = errors.New("wrong checksum reading data from flash")
var errFlashWrongLength = errors.New("unsupported flash data length")

const (
	FLASH_HEADER_BYTES       = 2 // checksum + length
	FLASH_GYR_CAL_BLOCKS     = 3 // gyro calibration offsets (int32 each)
	FLASH_GYR_CAL_BYTES      = FLASH_GYR_CAL_BLOCKS * 4
	FLASH_DEVICE_NAME_BYTES  = 16 // custom device name
	FLASH_AXIS_MAPPING_BYTES = 3  // axis mapping (3 bytes)
	FLASH_LENGTH             = FLASH_HEADER_BYTES + FLASH_GYR_CAL_BYTES + FLASH_DEVICE_NAME_BYTES + FLASH_AXIS_MAPPING_BYTES
)

type Flash struct {
	checksum      byte
	length        byte
	gyrCalOffsets [FLASH_GYR_CAL_BLOCKS]int32
	deviceName    [FLASH_DEVICE_NAME_BYTES]byte
	axisMapping   [FLASH_AXIS_MAPPING_BYTES]byte
}

func NewFlash() *Flash {
	return &Flash{
		checksum:      FLASH_LENGTH,
		length:        FLASH_LENGTH,
		gyrCalOffsets: [FLASH_GYR_CAL_BLOCKS]int32{0, 0, 0},
		deviceName:    [FLASH_DEVICE_NAME_BYTES]byte{'H', 'T'},
		axisMapping:   [FLASH_AXIS_MAPPING_BYTES]byte{0x10, 0x11, 0x12}, // default mapping: all axes enabled, not inverted, mapped to first 3 channels
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
	length := int(data[1])
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

	offset := FLASH_HEADER_BYTES

	// read gyro calibration, best effort
	if length < offset+FLASH_GYR_CAL_BYTES {
		println("Incomplete flash data, length:", length)
		return nil // this is fine, just no gyro calibration
	}
	for i := range FLASH_GYR_CAL_BLOCKS {
		fd.gyrCalOffsets[i] = toInt32(data[offset+i*4 : offset+(i+1)*4])
	}
	offset += FLASH_GYR_CAL_BYTES

	// read device name, best effort
	if length < offset+FLASH_DEVICE_NAME_BYTES {
		println("Incomplete flash data, length:", length)
		return nil // this is fine, just no name
	}
	for i := 0; i < FLASH_DEVICE_NAME_BYTES; i++ {
		fd.deviceName[i] = data[offset+i]
	}
	offset += FLASH_DEVICE_NAME_BYTES

	// read axis mapping, best effort
	if length < offset+FLASH_AXIS_MAPPING_BYTES {
		println("Incomplete flash data, length:", length)
		return nil // this is fine, just no mapping
	}
	for i := 0; i < FLASH_AXIS_MAPPING_BYTES; i++ {
		fd.axisMapping[i] = data[offset+i]
	}
	offset += FLASH_AXIS_MAPPING_BYTES

	println("Loaded from flash")
	println("  gyro calibration:", fd.gyrCalOffsets[0], fd.gyrCalOffsets[1], fd.gyrCalOffsets[2])
	println("  device name:", fd.DeviceName())
	println("  axis mapping:", fd.axisMapping[0], fd.axisMapping[1], fd.axisMapping[2])

	return nil
}

func (fd *Flash) Store() error {

	data := make([]byte, FLASH_LENGTH)

	data[1] = FLASH_LENGTH
	offset := FLASH_HEADER_BYTES

	// gyro calibration
	for i := range FLASH_GYR_CAL_BLOCKS {
		fromInt32(data[offset+i*4:offset+(i+1)*4], fd.gyrCalOffsets[i])
	}
	offset += FLASH_GYR_CAL_BYTES

	// device name
	for i := 0; i < FLASH_DEVICE_NAME_BYTES; i++ {
		data[offset+i] = fd.deviceName[i]
	}
	offset += FLASH_DEVICE_NAME_BYTES

	// axis mapping
	for i := 0; i < FLASH_AXIS_MAPPING_BYTES; i++ {
		data[offset+i] = fd.axisMapping[i]
	}
	offset += FLASH_AXIS_MAPPING_BYTES

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

	println("Stored to flash")
	println("  gyro calibration:", fd.gyrCalOffsets[0], fd.gyrCalOffsets[1], fd.gyrCalOffsets[2])
	println("  device name:", fd.DeviceName())
	println("  axis mapping:", fd.axisMapping[0], fd.axisMapping[1], fd.axisMapping[2])
	return nil
}

func (fd *Flash) SetGyrCalOffsets(offsets [FLASH_GYR_CAL_BLOCKS]int32, threshold int32) bool {
	overThreshold := false
	for i := range offsets {
		if abs(fd.gyrCalOffsets[i]-offsets[i]) > threshold {
			overThreshold = true
			break
		}
	}
	if overThreshold {
		fd.gyrCalOffsets = offsets
	}
	return overThreshold
}

func (fd *Flash) GyrCalOffsets() [FLASH_GYR_CAL_BLOCKS]int32 {
	return fd.gyrCalOffsets
}

func (fd *Flash) SetDeviceName(name string) bool {
	newName := false
	n := 0
	for n < min(FLASH_DEVICE_NAME_BYTES, len(name)) {
		if fd.deviceName[n] != name[n] {
			fd.deviceName[n] = name[n]
			newName = true
		}
		n++
	}
	for n < FLASH_DEVICE_NAME_BYTES {
		fd.deviceName[n] = 0
		n++
	}
	return newName
}

func (fd *Flash) DeviceName() string {
	n := 0
	for n < FLASH_DEVICE_NAME_BYTES && fd.deviceName[n] != 0 {
		n++
	}
	return string(fd.deviceName[:n])
}

func (fd *Flash) SetAxisMapping(mapping [FLASH_AXIS_MAPPING_BYTES]byte) bool {
	newMapping := false
	for i := 0; i < FLASH_AXIS_MAPPING_BYTES; i++ {
		if fd.axisMapping[i] != mapping[i] {
			fd.axisMapping[i] = mapping[i]
			newMapping = true
		}
	}
	return newMapping
}

func (fd *Flash) AxisMapping() [FLASH_AXIS_MAPPING_BYTES]byte {
	return fd.axisMapping
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
