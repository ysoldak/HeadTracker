module github.com/ysoldak/HeadTracker

go 1.19

require (
	github.com/go-gl/mathgl v1.0.0
	github.com/tracktum/go-ahrs v1.0.0
	tinygo.org/x/bluetooth v0.6.0
	tinygo.org/x/drivers v0.24.0
	tinygo.org/x/tinydraw v0.3.0
	tinygo.org/x/tinyfont v0.3.0
)

require (
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/muka/go-bluetooth v0.0.0-20220830075246-0746e3a1ea53 // indirect
	github.com/saltosystems/winrt-go v0.0.0-20230124093143-967a889c6c8f // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/tinygo-org/cbgo v0.0.4 // indirect
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
)

// replace tinygo.org/x/bluetooth => ../bluetooth
// replace tinygo.org/x/bluetooth => github.com/ysoldak/bluetooth sd-gatts-sys-attr-fix-panics
replace tinygo.org/x/bluetooth => github.com/ysoldak/bluetooth v0.3.1-0.20230425185611-fea9f9a95b8a
