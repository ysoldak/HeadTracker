package orientation

import (
	"time"

	"github.com/tracktum/go-ahrs"
)

type Orientation struct {
	imu              *IMU
	fusion           ahrs.Madgwick
	period           time.Duration
	Roll, Pitch, Yaw float32
}

func New() *Orientation {
	return &Orientation{
		imu: NewIMU(),
	}
}

func (o *Orientation) Configure(period time.Duration) {
	o.period = period
	o.imu.Configure()
	o.fusion = ahrs.NewMadgwick(0.01, float64(time.Second/o.period))
}

// TODO implement initial orientation when support magnetometer
// Init orientation (calculates and stores initial angles)
// func (o *Orientation) Init() {
// 	gx, gy, gz, ax, ay, az, err := o.imu.Read(true)
// 	q := orientationToQuaternion(ax, ay, az, 1, 0, 0) // assume N since we don't have mag
// 	initial[0], initial[1], initial[2] = quaternionToAngles(q)
// 	fusion.Quaternions = q
// }

// Update orientation
// TODO support magnetometer
func (o *Orientation) Update(fusion bool) {
	gx, gy, gz, ax, ay, az, err := o.imu.Read(true)
	if err != nil {
		println(err.Error())
		return
	}
	if fusion {
		q := o.fusion.Update6D(
			gx*degToRad, gy*degToRad, gz*degToRad,
			ax, ay, az,
		)
		o.Roll, o.Pitch, o.Yaw = quaternionToAngles(q)
	}
}

func (o *Orientation) Calibration() (roll, pitch, yaw int32) {
	return o.imu.gyrCal.offset[0], o.imu.gyrCal.offset[1], o.imu.gyrCal.offset[2]
}
