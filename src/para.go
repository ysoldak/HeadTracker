//go:build s140v7
// +build s140v7

package main

import (
	"machine"
	"time"

	"tinygo.org/x/bluetooth"
)

var blue = machine.LED_BLUE

var conStatusChan = make(chan bool, 1)

// var conTime time.Time

var ble = bluetooth.DefaultAdapter
var adv *bluetooth.Advertisement
var fff6Handle bluetooth.Characteristic

var channels = [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500}

// Have to send this to master radio on connect otherwise high chance opentx para code will never receive "Connected" message
// Since it looks for "Connected\r\n" and sometimes(?) bluetooth underlying layer on master radio
// never sends "\r\n" and starts sending trainer data directly
var bootBuffer = []byte{0x0d, 0x0a}

// 2 + 8(max 16) + 2
var paraBuffer []byte = make([]byte, 20)

// var paraBuffer = []byte{0x7e, 0x80, 0x85, 0x60, 0xa7, 0x55, 0x7d, 0x5d, 0xe5, 0xdc, 0x3d, 0xc3, 0xdc, 0x3d, 0xc3, 0x0f, 0x7e}

// Theory https://devzone.nordicsemi.com/f/nordic-q-a/15571/automatically-start-notification-upon-connection-event-manually-write-cccd---short-tutorial-on-notifications
// In practice these values were manually extracted after connecting to head tracker with BlueSee app
// That 0x01 out there is CCCD bit telling the bluetooth stack notification is enabled / client subscribed
// Last two bytes is CRC, see theory link
var fff6Attributes = []byte{0x0d, 0x00, 0x02, 0x00, 0x02, 0x00, 0x22, 0x00, 0x02, 0x00, 0x01, 0x00, 0xcd, 0xa0}

func paraSetup() {

	blue.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blue.High()

	must("enable BLE stack", ble.Enable())

	sysid := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.CharacteristicUUIDSystemID,
		Value:  []byte{0xF1, 0x63, 0x1B, 0xB0, 0x6F, 0x80, 0x28, 0xFE},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	manufacturer := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.CharacteristicUUIDManufacturerNameString,
		Value:  []byte{0x41, 0x70, 0x70},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	ieee := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.CharacteristicUUIDIEEE1107320601RegulatoryCertificationDataList,
		Value:  []byte{0xFE, 0x00, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6D, 0x65, 0x6E, 0x74, 0x61, 0x6C},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	pnpid := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.CharacteristicUUIDPnPID,
		Value:  []byte{0x01, 0x0D, 0x00, 0x00, 0x00, 0x10, 0x01},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	must("add service", ble.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDDeviceInformation,
		Characteristics: []bluetooth.CharacteristicConfig{
			sysid, manufacturer, ieee, pnpid,
		},
	}))

	fff1 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(0xFFF1),
		Value:  []byte{0x01},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
	}

	fff2 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(0xFFF2),
		Value:  []byte{0x02},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	fff3 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(0xFFF3),
		Value:  []byte{},
		Flags:  bluetooth.CharacteristicWriteWithoutResponsePermission,
	}

	fff5 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(0xFFF5),
		Value:  []byte{},
		Flags:  bluetooth.CharacteristicReadPermission,
	}

	fff6 := bluetooth.CharacteristicConfig{
		Handle: &fff6Handle,
		UUID:   bluetooth.New16BitUUID(0xFFF6),
		Value:  []byte{},
		Flags:  bluetooth.CharacteristicWriteWithoutResponsePermission | bluetooth.CharacteristicNotifyPermission,
	}

	must("add service", ble.AddService(&bluetooth.Service{
		UUID: bluetooth.New16BitUUID(0xFFF0),
		Characteristics: []bluetooth.CharacteristicConfig{
			fff1, fff2, fff3, fff5, fff6,
		},
	}))

	adv = ble.DefaultAdvertisement()
	must("config adv", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Hello",
		ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
	}))
	must("start adv", adv.Start())

	ble.SetConnectHandler(func(device bluetooth.Addresser, connected bool) {
		conStatusChan <- connected
	})

}

func paraWork(input chan [3]uint16) {
	connected := false
	address := ""
	var sendAfter time.Time
	for {
		select {
		case connected = <-conStatusChan:
			if connected {
				blue.Low()
				sendAfter = time.Now().Add(1 * time.Second)
				// conTime = time.Now()
				paraBoot()
				println("Connected")
			} else {
				blue.High()
				sendAfter = time.Time{}
				// conTime = time.Time{}
				println("Disconnected")
			}
			continue
		case values := <-input:
			channels[0] = values[0]
			channels[1] = values[1]
			channels[2] = values[2]
			continue
		default:
			if !connected || sendAfter.IsZero() || time.Now().Before(sendAfter) {
				if len(address) == 0 {
					addr, _ := ble.Address()
					address = addr.MAC.String()
				}
				println("Advertising as Hello /", address)
				time.Sleep(1000 * time.Millisecond)
				continue
			}
		}
		now := time.Now()
		paraSend()
		sleep := PERIOD - time.Since(now)
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

// paraBoot sends '\r\n', it helps remote switch to receiveTrainer state
func paraBoot() {
	fff6Handle.SetAttributes(fff6Attributes)
	fff6Handle.Write(bootBuffer)
}

func paraSet(idx byte, value uint16) {
	channels[idx] = value
}

const START_STOP byte = 0x7E
const BYTE_STUFF byte = 0x7D
const STUFF_MASK byte = 0x20

// Escapes bytes that equal to START_STOP and updates CRC
func paraPushByte(b byte, bufferIndex *byte, crc *byte) {
	*crc ^= b
	if b == START_STOP || b == BYTE_STUFF {
		paraBuffer[*bufferIndex] = BYTE_STUFF
		*bufferIndex++
		b ^= STUFF_MASK
	}
	paraBuffer[*bufferIndex] = b
	*bufferIndex++
}

// Encodes channels array to para trainer packet (adapted from OpenTX source code)
func paraSend() {

	var bufferIndex byte = 0
	var crc byte = 0x00

	paraBuffer[bufferIndex] = START_STOP
	bufferIndex++
	paraPushByte(0x80, &bufferIndex, &crc)
	for channel := 0; channel < 8; channel += 2 {
		channelValue1 := channels[channel]
		channelValue2 := channels[channel+1]
		paraPushByte(byte(channelValue1&0x00ff), &bufferIndex, &crc)
		paraPushByte(byte((channelValue1&0x0f00)>>4)+byte((channelValue2&0x00f0)>>4), &bufferIndex, &crc)
		paraPushByte(byte((channelValue2&0x000f)<<4)+byte((channelValue2&0x0f00)>>8), &bufferIndex, &crc)
	}
	paraBuffer[bufferIndex] = crc
	paraBuffer[bufferIndex+1] = START_STOP

	// fmt.Printf("%x\r\n", paraBuffer[:bufferIndex+2])

	fff6Handle.Write(paraBuffer[:bufferIndex+2])
}
