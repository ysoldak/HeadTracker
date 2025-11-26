package trainer

// Bluetooth (FrSKY's PARA trainer protocol) link

import (
	"time"

	"tinygo.org/x/bluetooth"
)

const STATE_DEVICE_NAME = 0xFF01

const RUNTIME_COMMANDS = 0xFF41
const RUNTIME_COMMANDS_COMPATIBILITY = 0xAFF2 // Cliff's Head Tracker reset command characteristic
const RUNTIME_ORIENTATION_RESET = 'R'
const RUNTIME_FACTORY_RESET = 'F'

// Have to send this to master radio on connect otherwise high chance opentx para code will never receive "Connected" message
// Since it looks for "Connected\r\n" and sometimes(?) bluetooth underlying layer on master radio
// never sends "\r\n" and starts sending trainer data directly
var bootBuffer = []byte{0x0d, 0x0a}

type Para struct {
	adapter    *bluetooth.Adapter
	adv        *bluetooth.Advertisement
	fff6Handle bluetooth.Characteristic

	buffer    [20]byte
	sendAfter time.Time

	paired   bool
	address  string
	channels [8]uint16

	orientationReset bool
	factoryReset     bool

	nameChanged bool
	nameBytes   [16]byte // allocations are not allowed in bluetooth event handlers, so working with byte array
}

func NewPara() *Para {
	return &Para{
		adapter:  bluetooth.DefaultAdapter,
		paired:   false,
		address:  "B1:6B:00:B5:BA:BE",
		channels: [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
	}
}

func (t *Para) Configure(name string) {
	n := 0
	for n < len(name) && n < len(t.nameBytes) {
		t.nameBytes[n] = name[n]
		n++
	}
	for n < len(t.nameBytes) {
		t.nameBytes[n] = 0
		n++
	}

	t.adapter.Enable()

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

	t.adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.ServiceUUIDDeviceInformation,
		Characteristics: []bluetooth.CharacteristicConfig{
			sysid, manufacturer, ieee, pnpid,
		},
	})

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
		Handle: &t.fff6Handle,
		UUID:   bluetooth.New16BitUUID(0xFFF6),
		Value:  []byte{},
		Flags:  bluetooth.CharacteristicWriteWithoutResponsePermission | bluetooth.CharacteristicNotifyPermission,
	}

	// Commands
	// R - reset orientation, compatible with Cliff's HT reset button
	commandHandler := func(client bluetooth.Connection, offset int, value []byte) {
		if len(value) == 1 && value[0] == RUNTIME_ORIENTATION_RESET {
			t.orientationReset = true
		}
		if len(value) == 1 && value[0] == RUNTIME_FACTORY_RESET {
			t.factoryReset = true
		}
	}
	ff41 := bluetooth.CharacteristicConfig{
		Handle:     nil,
		UUID:       bluetooth.New16BitUUID(RUNTIME_COMMANDS),
		Value:      []byte{},
		Flags:      bluetooth.CharacteristicWritePermission,
		WriteEvent: commandHandler,
	}
	// Compatibility with Cliff's Head Tracker
	aff2 := bluetooth.CharacteristicConfig{
		Handle:     nil,
		UUID:       bluetooth.New16BitUUID(RUNTIME_COMMANDS_COMPATIBILITY),
		Value:      []byte{},
		Flags:      bluetooth.CharacteristicWritePermission,
		WriteEvent: commandHandler,
	}

	// Device Name
	ff01 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(STATE_DEVICE_NAME),
		Value:  []byte(name), // FIXME: this may become larger than 16 bytes if client writes more data
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) == 0 {
				return
			}
			n := 0
			for n < len(t.nameBytes) && n < len(value) {
				t.nameBytes[n] = value[n]
				n++
			}
			for n < len(t.nameBytes) {
				t.nameBytes[n] = 0
				n++
			}
			t.nameChanged = true
		},
	}

	t.adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.New16BitUUID(0xFFF0),
		Characteristics: []bluetooth.CharacteristicConfig{
			fff1,
			fff2,
			fff3,
			fff5,
			fff6, // channels data
			ff01, // device name
			ff41, // runtime commands from client
			aff2, // for compatibility with Cliff's HT
		},
	})

	t.adv = t.adapter.DefaultAdvertisement()
	t.adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    string(t.nameBytes[:]),
		ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
	})
	t.adv.Start()

	addr, _ := t.adapter.Address()
	t.address = addr.MAC.String()

	t.adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if connected {
			t.sendAfter = time.Now().Add(1 * time.Second) // wait for 1 second before sending data
			setSoftDeviceSystemAttributes()               // force enable notify for fff6
			t.fff6Handle.Write(bootBuffer)                // send '\r\n', it helps remote master switch to receiveTrainer state
			t.paired = true
		} else {
			t.paired = false
		}
	})

}

func (t *Para) Start() {
	// no-op
}

func (t *Para) Update() {
	if !t.paired {
		return
	}
	if !time.Now().After(t.sendAfter) {
		return
	}
	size := t.encode()
	n, err := t.fff6Handle.Write(t.buffer[:size])
	if err != nil {
		println(err.Error())
		println(n)
	}
}

func (p *Para) Paired() bool {
	return p.paired
}

func (p *Para) Address() string {
	return p.address
}

func (p *Para) Channels() []uint16 {
	return p.channels[:3]
}

func (p *Para) SetChannel(n int, v uint16) {
	p.channels[n] = v
}

func (p *Para) OrientationReset() bool {
	if p.orientationReset {
		p.orientationReset = false
		return true
	}
	return false
}

func (p *Para) FactoryReset() bool {
	if p.factoryReset {
		p.factoryReset = false
		return true
	}
	return false
}

func (p *Para) Name() (string, bool) {
	if p.nameChanged {
		p.nameChanged = false
		// update advertisement name only when changed
		// can't do it in the write handler due to allocation restrictions
		// we expect Name() to be called frequently in the main loop
		p.adv.Configure(bluetooth.AdvertisementOptions{
			LocalName:    string(p.nameBytes[:]),
			ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
		})
		return string(p.nameBytes[:]), true
	}
	return "", false
}

// -- PARA Protocol ------------------------------------------------------------

// 2 + 8(max 16) + 2

const START_STOP byte = 0x7E
const BYTE_STUFF byte = 0x7D
const STUFF_MASK byte = 0x20

// Escapes bytes that equal to START_STOP and updates CRC
func (t *Para) push(b byte, bufferIndex *byte, crc *byte) {
	*crc ^= b
	if b == START_STOP || b == BYTE_STUFF {
		t.buffer[*bufferIndex] = BYTE_STUFF
		*bufferIndex++
		b ^= STUFF_MASK
	}
	t.buffer[*bufferIndex] = b
	*bufferIndex++
}

// Encodes the channels array to a para trainer packet (adapted from OpenTX source code)
func (t *Para) encode() byte {

	var bufferIndex byte = 0
	var crc byte = 0x00

	t.buffer[bufferIndex] = START_STOP
	bufferIndex++
	t.push(0x80, &bufferIndex, &crc)
	for channel := 0; channel < 8; channel += 2 {
		channelValue1 := t.channels[channel]
		channelValue2 := t.channels[channel+1]
		t.push(byte(channelValue1&0x00ff), &bufferIndex, &crc)
		t.push(byte((channelValue1&0x0f00)>>4)+byte((channelValue2&0x00f0)>>4), &bufferIndex, &crc)
		t.push(byte((channelValue2&0x000f)<<4)+byte((channelValue2&0x0f00)>>8), &bufferIndex, &crc)
	}
	t.buffer[bufferIndex] = crc
	t.buffer[bufferIndex+1] = START_STOP

	return bufferIndex + 2
}
