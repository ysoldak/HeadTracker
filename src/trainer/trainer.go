package trainer

type Trainer interface {
	Configure()
	Run()
	Paired() bool
	Address() string
	Channels() []uint16
	SetChannel(num int, v uint16)
}
