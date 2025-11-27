package trainer

// Bluetooth (FrSKY's PARA trainer protocol) link

import (
	"time"

	"tinygo.org/x/bluetooth"
)

const CHAR_COMMANDS = 0xFFC1        // 'C' is for Commands (runtime effects)
const CHAR_COMMANDS_COMPAT = 0xAFF2 // Cliff's Head Tracker reset command characteristic
const CMD_ORIENTATION_RESET = 'R'
const CMD_FACTORY_RESET = 'F'
const CMD_REBOOT = 'B'

const CHAR_DATA_DEVICE_NAME = 0xFFD1 // 'D' is for Data (to persist)
const CHAR_DATA_CH_OFFSET = 0xFFD2
const CHAR_DATA_CH_ORDER = 0xFFD3
const CHAR_DATA_CH_ENABLED = 0xFFD4

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

	sysHandler SystemHandler

	orientationReset bool
	factoryReset     bool
	reboot           bool
	name             struct {
		changed bool
		len     int
		buffer  [16]byte // allocations are not allowed in bluetooth event handlers, so working with byte array
	}
	channelsConfig struct {
		changed bool
		offset  byte
		order   byte
		enabled byte
	}
}

// Communication with the system (main application)
type SystemHandler interface {
	OrientationReset()
	FactoryReset()
	Reboot()

	Name() string
	SetName(newName string)

	ChannelsConfig() (byte, byte, byte)
	SetChannelsConfig(offset byte, order byte, enabled byte)
}

func NewPara(sysHandler SystemHandler) *Para {
	return &Para{
		adapter:    bluetooth.DefaultAdapter,
		paired:     false,
		address:    "B1:6B:00:B5:BA:BE",
		channels:   [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
		sysHandler: sysHandler,
	}
}

func (t *Para) Configure() {
	name := t.sysHandler.Name()
	n := 0
	for n < len(name) && n < len(t.name.buffer) {
		t.name.buffer[n] = name[n]
		n++
	}
	t.name.len = n

	t.channelsConfig.offset, t.channelsConfig.order, t.channelsConfig.enabled = t.sysHandler.ChannelsConfig()

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
		if len(value) == 1 && value[0] == CMD_ORIENTATION_RESET {
			t.orientationReset = true
		}
		if len(value) == 1 && value[0] == CMD_FACTORY_RESET {
			t.factoryReset = true
		}
		if len(value) == 1 && value[0] == CMD_REBOOT {
			t.reboot = true
		}
	}
	charCmd := bluetooth.CharacteristicConfig{
		Handle:     nil,
		UUID:       bluetooth.New16BitUUID(CHAR_COMMANDS),
		Value:      []byte{},
		Flags:      bluetooth.CharacteristicWritePermission,
		WriteEvent: commandHandler,
	}
	// Compatibility with Cliff's Head Tracker
	charCmdCompat := bluetooth.CharacteristicConfig{
		Handle:     nil,
		UUID:       bluetooth.New16BitUUID(CHAR_COMMANDS_COMPAT),
		Value:      []byte{},
		Flags:      bluetooth.CharacteristicWritePermission,
		WriteEvent: commandHandler,
	}

	// Device Name
	charDataDeviceName := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_DEVICE_NAME),
		Value:  []byte(name), // FIXME: this may become larger than 16 bytes if client writes more data
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) == 0 {
				return
			}
			n := 0
			for n < len(t.name.buffer) && n < len(value) {
				t.name.buffer[n] = value[n]
				n++
			}
			t.name.len = n
			t.name.changed = true
		},
	}

	charChOffset := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_CH_OFFSET),
		Value:  []byte{t.channelsConfig.offset},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.channelsConfig.offset = value[0]
			t.channelsConfig.changed = true
		},
	}

	charChOrder := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_CH_ORDER),
		Value:  []byte{t.channelsConfig.order},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.channelsConfig.order = value[0]
			t.channelsConfig.changed = true
		},
	}

	chEnabled := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_CH_ENABLED),
		Value:  []byte{t.channelsConfig.enabled},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.channelsConfig.enabled = value[0]
			t.channelsConfig.changed = true
		},
	}

	t.adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.New16BitUUID(0xFFF0),
		Characteristics: []bluetooth.CharacteristicConfig{
			fff1,
			fff2,
			fff3,
			fff5,
			fff6,               // channels data
			charCmd,            // runtime commands from client
			charCmdCompat,      // for compatibility with Cliff's HT
			charDataDeviceName, // device name
			charChOffset,
			charChOrder,
			chEnabled,
		},
	})

	t.adv = t.adapter.DefaultAdvertisement()
	t.adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    string(t.name.buffer[:t.name.len]),
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
	// send channels data
	size := t.encode()
	n, err := t.fff6Handle.Write(t.buffer[:size])
	if err != nil {
		println(err.Error())
		println(n)
	}
	// handle commands
	if t.orientationReset {
		t.sysHandler.OrientationReset()
		t.orientationReset = false
	}
	if t.factoryReset {
		t.sysHandler.FactoryReset()
		t.factoryReset = false
	}
	if t.reboot {
		t.sysHandler.Reboot()
		t.reboot = false
	}
	if t.name.changed {
		t.sysHandler.SetName(string(t.name.buffer[:t.name.len]))
		// update advertisement name
		// can't do it in the write handler or on disconnect due to allocation restrictions in interrupt handlers
		t.adv.Configure(bluetooth.AdvertisementOptions{
			LocalName:    t.sysHandler.Name(),
			ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
		})
		t.name.changed = false
	}
	if t.channelsConfig.changed {
		t.sysHandler.SetChannelsConfig(t.channelsConfig.offset, t.channelsConfig.order, t.channelsConfig.enabled)
		t.channelsConfig.changed = false
	}
}

func (p *Para) Paired() bool {
	return p.paired
}

func (p *Para) Address() string {
	return p.address
}

func (p *Para) SetChannel(n int, v uint16) {
	p.channels[n] = v
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
