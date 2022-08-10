package trainer

type Trainer interface {
	Configure()
	Run()
	Paired() bool
	Address() string
	Channels() [8]uint16
	SetChannel(num int, v uint16)
}
