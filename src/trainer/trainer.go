package trainer

type Trainer interface {
	Configure(name string)
	Start()
	Update()
	Paired() bool
	Address() string
	Channels() []uint16
	SetChannel(num int, v uint16)
}
