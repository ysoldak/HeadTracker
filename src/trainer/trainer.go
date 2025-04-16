package trainer

type Trainer interface {
	Configure()
	Start()
	Update()
	Paired() bool
	Address() string
	Channels() []uint16
	SetChannel(num int, v uint16)
}
