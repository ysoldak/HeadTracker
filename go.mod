module github.com/ysoldak/HeadTracker

go 1.17

require (
	github.com/go-gl/mathgl v1.0.0
	github.com/tracktum/go-ahrs v1.0.0
	tinygo.org/x/bluetooth v0.5.0
	tinygo.org/x/drivers v0.21.0
	tinygo.org/x/tinydraw v0.0.0-20211229235716-f3521ee65ebb
	tinygo.org/x/tinyfont v0.2.1
)

require (
	github.com/JuulLabs-OSS/cbgo v0.0.2 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/muka/go-bluetooth v0.0.0-20220323170840-382ca1d29f29 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	golang.org/x/image v0.0.0-20190321063152-3fc05d484e9f // indirect
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da // indirect
)

replace tinygo.org/x/bluetooth => ../bluetooth // github.com/ysoldak/bluetooth sd-gatts-sys-attr
