package trainer

type Trainer interface {
	Configure(name string)
	Start()
	Channels() []uint16
	SetChannel(num int, v uint16)

	// Bluetooth specific
	Update()
	Paired() bool
	Address() string
	ResetRequested() bool
}
