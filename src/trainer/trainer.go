package trainer

type Trainer interface {
	Configure()
	Start()
	SetChannel(num int, v uint16)

	// Bluetooth specific
	Update()
	Paired() bool
	Address() string
}
