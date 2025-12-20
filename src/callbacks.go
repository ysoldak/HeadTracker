package main

type BluetoothCallbackHandler struct {
}

func (b *BluetoothCallbackHandler) OnConnect() {
	state.connected = true
	println("Bluetooth connected")
}

func (b *BluetoothCallbackHandler) OnDisconnect() {
	state.connected = false
	println("Bluetooth disconnected")
}

func (b *BluetoothCallbackHandler) OnOrientationReset() {
	o.Reset()
	println("Orientation reset via Bluetooth command")
}
