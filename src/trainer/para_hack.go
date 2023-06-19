package trainer

// #include "ble.h"
import "C"

// Theory https://devzone.nordicsemi.com/f/nordic-q-a/15571/automatically-start-notification-upon-connection-event-manually-write-cccd---short-tutorial-on-notifications
// In practice these values were manually extracted after connecting to head tracker with BlueSee app
// That 0x01 out there is CCCD bit telling the bluetooth stack notification is enabled / client subscribed
// Last two bytes is CRC, see the theory link.
var sysAttributes = []byte{0x0d, 0x00, 0x02, 0x00, 0x02, 0x00, 0x22, 0x00, 0x02, 0x00, 0x01, 0x00, 0xcd, 0xa0}

// setSystemAttributes including CCCD notification bit for FFF6 telling the bluetooth stack notification is enabled / client subscribed.
// Note: the bluetooth package does not export the active connection handle, so trying different handles sequentially until success.
func setSoftDeviceSystemAttributes() {
	length := uint16(len(sysAttributes))
	connHandle := uint16(1)
	for {
		err := C.sd_ble_gatts_sys_attr_set(connHandle, &sysAttributes[0], length, 0)
		if err == 0x0 { // NRF_SUCCESS
			return
		}
		if err == 0x3002 { // BLE_ERROR_INVALID_CONN_HANDLE
			connHandle++
		}
		if connHandle > 128 {
			return
		}
	}
}
