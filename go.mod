module github.com/ysoldak/HeadTracker

go 1.17

require tinygo.org/x/bluetooth v0.3.0

require github.com/ysoldak/magcal v0.1.1

require tinygo.org/x/drivers v0.18.0

require github.com/tracktum/go-ahrs v1.0.0

require (
	github.com/JuulLabs-OSS/cbgo v0.0.2 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/muka/go-bluetooth v0.0.0-20210812063148-b6c83362e27d // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da // indirect
)

replace tinygo.org/x/bluetooth => ../bluetooth

replace github.com/ysoldak/magcal => ../magcal
