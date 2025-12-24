package trainer

// Bluetooth (FrSKY's PARA trainer protocol) link

import (
	"time"

	"tinygo.org/x/bluetooth"
)

// 'C' is for Commands (runtime effects)
const (
	CHAR_COMMANDS         = 0xFFC1
	CHAR_COMMANDS_COMPAT  = 0xAFF2 // Cliff's Head Tracker reset command characteristic
	CMD_ORIENTATION_RESET = 'R'    // reset orientation
	CMD_FACTORY_RESET     = 'F'    // factory reset
	CMD_REBOOT            = 'B'    // reboot device
)

// 'D' is for Data (to persist)
const (
	CHAR_DATA_AXIS1 = 0xFFD1 // axis 1 mapping (0x00 to 0x0F and 0xFF, anything else is ignored)
	CHAR_DATA_AXIS2 = 0xFFD2 // axis 2 mapping (to invert, add 0x08: 0x0A is the same as 0x02 but inverted)
	CHAR_DATA_AXIS3 = 0xFFD3 // axis 3 mapping (0xFF disables the axis mapping)
)

// Have to send this to master radio on connect otherwise high chance opentx para code will never receive "Connected" message
// Since it looks for "Connected\r\n" and sometimes(?) bluetooth underlying layer on master radio
// never sends "\r\n" and starts sending trainer data directly
var bootBuffer = []byte{0x0d, 0x0a}

type Para struct {
	adapter    *bluetooth.Adapter
	adv        *bluetooth.Advertisement
	fff6Handle bluetooth.Characteristic

	name            string
	callbackHandler CallbackHandler

	buffer    [20]byte
	sendAfter time.Time

	paired   bool
	channels [8]uint16

	orientationReset bool
	factoryReset     bool
	reboot           bool

	data struct {
		axisValues  [3]byte
		axisChanged [3]bool
	}
}

type CallbackHandler interface {
	// bluetooth connection events
	OnConnect()
	OnDisconnect()

	// remote commands
	OnOrientationReset()
	OnReboot()
	OnFactoryReset()

	// remote configuration
	OnAxisMappingChange(axis byte, value byte)
}

func NewPara(name string, callbackHandler CallbackHandler) *Para {
	return &Para{
		adapter:         bluetooth.DefaultAdapter,
		name:            name,
		callbackHandler: callbackHandler,
		paired:          false,
		channels:        [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
	}
}

func (t *Para) Start() string {
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
	// F - factory reset
	// B - reboot device
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

	charDataAxis1 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_AXIS1),
		Value:  []byte{t.data.axisValues[0]},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.data.axisValues[0] = value[0]
			t.data.axisChanged[0] = true
		},
	}
	charDataAxis2 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_AXIS2),
		Value:  []byte{t.data.axisValues[1]},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.data.axisValues[1] = value[0]
			t.data.axisChanged[1] = true
		},
	}
	charDataAxis3 := bluetooth.CharacteristicConfig{
		Handle: nil,
		UUID:   bluetooth.New16BitUUID(CHAR_DATA_AXIS3),
		Value:  []byte{t.data.axisValues[2]},
		Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			if len(value) != 1 {
				return
			}
			t.data.axisValues[2] = value[0]
			t.data.axisChanged[2] = true
		},
	}

	t.adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.New16BitUUID(0xFFF0),
		Characteristics: []bluetooth.CharacteristicConfig{
			fff1,
			fff2,
			fff3,
			fff5,
			fff6,          // channels data
			charCmd,       // runtime commands from client
			charCmdCompat, // for compatibility with Cliff's HT
			charDataAxis1, // axis 1 mapping
			charDataAxis2, // axis 2 mapping
			charDataAxis3, // axis 3 mapping
		},
	})

	t.adv = t.adapter.DefaultAdvertisement()
	t.adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    t.name,
		ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
	})
	t.adv.Start()

	t.adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if connected {
			t.sendAfter = time.Now().Add(1 * time.Second) // wait for 1 second before sending data
			setSoftDeviceSystemAttributes()               // force enable notify for fff6
			t.fff6Handle.Write(bootBuffer)                // send '\r\n', it helps remote master switch to receiveTrainer state
			t.paired = true
			t.callbackHandler.OnConnect()
		} else {
			t.paired = false
			t.callbackHandler.OnDisconnect()
		}
	})

	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		for range ticker.C {

			t.update()

			if t.orientationReset {
				t.orientationReset = false
				t.callbackHandler.OnOrientationReset()
			}
			if t.factoryReset {
				t.factoryReset = false
				t.callbackHandler.OnFactoryReset()
			}
			if t.reboot {
				t.reboot = false
				t.callbackHandler.OnReboot()
			}

			for i := byte(0); i < 3; i++ {
				if t.data.axisChanged[i] {
					t.data.axisChanged[i] = false
					// sanity check for invalid values
					if t.data.axisValues[i] > 0x0F && t.data.axisValues[i] != 0xFF {
						continue
					}
					t.callbackHandler.OnAxisMappingChange(i, t.data.axisValues[i])
				}
			}

		}
	}()

	addr, _ := t.adapter.Address()
	return addr.MAC.String()

}

func (t *Para) update() {
	if !t.paired {
		return
	}
	if !time.Now().After(t.sendAfter) {
		return
	}
	size := t.encode()
	n, err := t.fff6Handle.Write(t.buffer[:size])
	if err != nil {
		println("FFF6 write error:", err.Error(), n)
	}
}

func (p *Para) SetChannel(n byte, v uint16) {
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
