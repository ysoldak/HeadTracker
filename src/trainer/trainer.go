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

	// Remote controls
	OrientationReset() bool // whether an orientation reset was requested
	FactoryReset() bool     // whether a factory reset was requested
	Name() (string, bool)   // new name and whether it changed
}
