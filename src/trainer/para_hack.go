package trainer

/*
// Define SoftDevice functions as regular function declarations (not inline static functions).
#define SVCALL_AS_NORMAL_FUNCTION
#include "ble.h"
*/
import "C"

// Theory https://devzone.nordicsemi.com/f/nordic-q-a/15571/automatically-start-notification-upon-connection-event-manually-write-cccd---short-tutorial-on-notifications
// In practice these values were manually extracted after connecting to head tracker with BlueSee app
// That 0x01 out there is CCCD bit telling the bluetooth stack notification is enabled / client subscribed
// Last two bytes is CRC, see theory link
var sysAttributes = []byte{0x0d, 0x00, 0x02, 0x00, 0x02, 0x00, 0x22, 0x00, 0x02, 0x00, 0x01, 0x00, 0xcd, 0xa0}

// setSystemAttributes including CCCD notification bit for FFF6 telling the bluetooth stack notification is enabled / client subscribed.
// The tricky part here is to get the connection handle since it's not exported by the bluetooth package, so trying set until success.
// Assumption here is we have only one active connection at any time.
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
