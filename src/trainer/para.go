package trainer

// Bluetooth (FrSKY's PARA trainer protocol) link

import (
	"time"

	"tinygo.org/x/bluetooth"
)

// Have to send this to master radio on connect otherwise high chance opentx para code will never receive "Connected" message
// Since it looks for "Connected\r\n" and sometimes(?) bluetooth underlying layer on master radio
// never sends "\r\n" and starts sending trainer data directly
var bootBuffer = []byte{0x0d, 0x0a}

type Para struct {
	adapter    *bluetooth.Adapter
	adv        *bluetooth.Advertisement
	fff6Handle bluetooth.Characteristic

	buffer    []byte
	sendDelay time.Duration

	paired   bool
	address  string
	channels [8]uint16
}

func NewPara() *Para {
	return &Para{
		adapter:  bluetooth.DefaultAdapter,
		buffer:   make([]byte, 20),
		paired:   false,
		address:  "B1:6B:00:B5:BA:BE",
		channels: [8]uint16{1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500},
	}
}

func (t *Para) Configure() {
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

	t.adapter.AddService(&bluetooth.Service{
		UUID: bluetooth.New16BitUUID(0xFFF0),
		Characteristics: []bluetooth.CharacteristicConfig{
			fff1, fff2, fff3, fff5, fff6,
		},
	})

	t.adv = t.adapter.DefaultAdvertisement()
	t.adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Hello",
		ServiceUUIDs: []bluetooth.UUID{bluetooth.New16BitUUID(0xFFF0)},
	})
	t.adv.Start()

	addr, _ := t.adapter.Address()
	t.address = addr.MAC.String()

	t.adapter.SetConnectHandler(func(device bluetooth.Address, connected bool) {
		if connected {
			t.sendDelay = 1 * time.Second
			t.paired = true
		} else {
			t.paired = false
		}
	})

}

func (t *Para) Run() {
	period := 20 * time.Millisecond
	for {
		time.Sleep(period)
		if !t.paired {
			continue
		}
		if t.sendDelay > 0 {
			if t.sendDelay == 1*time.Second {
				setSoftDeviceSystemAttributes() // force enable notify for fff6
				t.fff6Handle.Write(bootBuffer)  // send '\r\n', it helps remote master switch to receiveTrainer state
			}
			t.sendDelay -= period
			continue
		}
		size := t.encode()
		n, err := t.fff6Handle.Write(t.buffer[:size])
		if err != nil {
			println(err.Error())
			println(n)
		}
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
