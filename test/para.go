package main

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const START_STOP byte = 0x7E
const BYTE_STUFF byte = 0x7D
const STUFF_MASK byte = 0x20

const dataTimeout = 5 * time.Second

type ParaTrainer struct {
	adapter *bluetooth.Adapter
	device  bluetooth.Device

	address *bluetooth.Address

	Connected bool
	Channels  []uint16

	char     bluetooth.DeviceCharacteristic
	dataCh   chan []byte
	dataTime time.Time // force disconnect if no data received after this timestamp

	callbackConnected    func()
	callbackDisconnected func()
}

func NewParaTrainer() *ParaTrainer {
	pt := ParaTrainer{
		adapter:   bluetooth.DefaultAdapter,
		Connected: false,
		Channels:  make([]uint16, 8),
		dataCh:    make(chan []byte, 100),
	}
	pt.adapter.Enable()
	time.Sleep(1 * time.Second)
	return &pt
}

func (pt *ParaTrainer) connect() (err error) {

	fmt.Println("ParaTrainer: Connecting to", *pt.address)

	pt.device, err = pt.adapter.Connect(*pt.address, bluetooth.ConnectionParams{})
	if err != nil {
		return fmt.Errorf("failed to connect: %s", err)
	}

	srvcs, err := pt.device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)})
	if err != nil {
		return fmt.Errorf("failed to discover services: %s", err)
	}

	if len(srvcs) == 0 {
		return fmt.Errorf("could not find head tracking service")
	}

	srvc := srvcs[0]

	chars, err := srvc.DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0xFFF6)})
	if err != nil {
		return fmt.Errorf("failed to discover characteristics: %s", err)
	}

	if len(chars) == 0 {
		return fmt.Errorf("could not find head tracking characteristic")
	}

	pt.char = chars[0]
	pt.char.EnableNotifications(func(buf []byte) {
		// d2Pin.High()
		select {
		case pt.dataCh <- buf:
		default:
			fmt.Println("ParaTrainer: Packet drop")
		}
		// d2Pin.Low()
	})

	pt.dataTime = time.Now().Add(dataTimeout)

	pt.Connected = true

	fmt.Printf("ParaTrainer: Connected to %s\n", pt.address.String())

	return nil
}

func (pt *ParaTrainer) Scan() (err error) {

	// time.Sleep(5 * time.Second)

	for {
		time.Sleep(1 * time.Second)
		if pt.Connected {
			continue
		}
		fmt.Println("ParaTrainer: Start scanning for devices")
		err = pt.adapter.Scan(pt.doScan)
		if err != nil {
			fmt.Printf("ParaTrainer: Failed to scan: %s\r\n", err)
			pt.adapter.StopScan()
		}
		if pt.address != nil {
			err = pt.connect()
			if err != nil {
				fmt.Printf("ParaTrainer: Failed to connect: %s\r\n", err)
			}
			if pt.callbackConnected != nil {
				pt.callbackConnected()
			}
		}
	}
}

func (pt *ParaTrainer) doScan(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
	if result.LocalName() == "Hello" {
		fmt.Printf("ParaTrainer: Found device %s\r\n", result.Address.String())
		pt.address = &result.Address
		adapter.StopScan()
	}
}

// Read out the data channel and process the most recent packet.
// This guarantees no clash between reading and writing.
func (pt *ParaTrainer) Read() {
	// d1Pin.High()
	// defer d1Pin.Low()
	var buf []byte = nil
	stop := false
	for !stop {
		select {
		case buf = <-pt.dataCh:
			// nop
		default:
			stop = true
		}
	}
	if buf != nil {
		channels, err := pt.decode(buf)
		if err != nil {
			fmt.Printf("ParaTrainer: Failed to decode: %s\r\n", err)
			return
		}
		for i := 0; i < 8; i++ {
			pt.Channels[i] = channels[i]
		}
		pt.dataTime = time.Now().Add(dataTimeout)
		return
	}
	if pt.Connected && (time.Now().After(pt.dataTime)) {
		fmt.Println("ParaTrainer: Data Timeout")
		fmt.Println("ParaTrainer: Disconnecting")
		if pt.callbackDisconnected != nil {
			pt.callbackDisconnected()
		}
		pt.char.EnableNotifications(nil)
		go pt.device.Disconnect() // this may take some time, so do it in a separate thread
		pt.Connected = false
		pt.address = nil
		fmt.Println("ParaTrainer: Disconnected")
		return
	}
}

// Decodes a para trainer packet into the channels array
func (pt *ParaTrainer) decode(buffer []byte) (channels []uint16, err error) {

	channels = make([]uint16, 8)

	var b byte
	var bufferIndex byte = 0
	var crc byte = 0x00

	if buffer[bufferIndex] != START_STOP {
		return nil, fmt.Errorf("invalid start byte")
	}
	bufferIndex++
	b, bufferIndex = pop(buffer, &bufferIndex, &crc)
	if b != 0x80 {
		return nil, fmt.Errorf("invalid packet type")
	}

	for channel := 0; channel < 8; channel += 2 {
		b, bufferIndex = pop(buffer, &bufferIndex, &crc)
		channelValue1 := uint16(b)

		b, bufferIndex = pop(buffer, &bufferIndex, &crc)
		channelValue1 |= uint16(b&0xF0) << 4
		channelValue2 := uint16(b&0x0F) << 4

		b, bufferIndex = pop(buffer, &bufferIndex, &crc)
		channelValue2 |= uint16(b&0x0F)<<8 | uint16(b&0xF0)>>4

		channels[channel] = channelValue1
		channels[channel+1] = channelValue2
	}

	if buffer[bufferIndex] != crc {
		fmt.Printf("CRC mismatch: %X != %X\n", buffer[bufferIndex], crc)
		return nil, fmt.Errorf("CRC mismatch")
	}
	if buffer[bufferIndex+1] != START_STOP {
		return nil, fmt.Errorf("invalid stop byte")
	}

	return channels, nil
}

// Unescapes bytes that were stuffed and updates CRC
func pop(buffer []byte, bufferIndex *byte, crc *byte) (byte, byte) {
	b := buffer[*bufferIndex]
	*bufferIndex++
	if b == BYTE_STUFF {
		b = buffer[*bufferIndex] ^ STUFF_MASK
		*bufferIndex++
	}
	*crc ^= b
	return b, *bufferIndex
}
