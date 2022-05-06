package orientation

import (
	"time"

	"github.com/tracktum/go-ahrs"
)

type Orientation struct {
	imu    *IMU
	fusion ahrs.Madgwick
	angles [2][3]float32 // initial and current, 3 of each
}

func New() *Orientation {
	return &Orientation{
		imu: NewIMU(),
	}
}

func (o *Orientation) Configure(period time.Duration) {
	o.imu.Configure()
	o.fusion = ahrs.NewMadgwick(0.025, float64(time.Second/period))
}

// TODO implement initial orientation when support magnetometer
// Init orientation (calculates and stores initial angles)
func (o *Orientation) Center() {
	_, _, _, ax, ay, az, _ := o.imu.Read()
	q := orientationToQuaternion(ax, ay, az, 1, 0, 0) // assume N since we don't have mag
	o.angles[0][0], o.angles[0][1], o.angles[0][2] = quaternionToAngles(q)
	o.fusion.Quaternions = q
}

// Update orientation
// TODO support magnetometer
func (o *Orientation) Update(fusion bool) {
	gx, gy, gz, ax, ay, az, err := o.imu.Read()
	if err != nil {
		println(err.Error())
		return
	}
	if fusion {
		q := o.fusion.Update6D(
			gx*degToRad, gy*degToRad, gz*degToRad,
			ax, ay, az,
		)
		o.angles[1][0], o.angles[1][1], o.angles[1][2] = quaternionToAngles(q)
		o.angles[1][0] -= o.angles[0][0]
		o.angles[1][1] -= o.angles[0][1]
		o.angles[1][2] -= o.angles[0][2]
	}
}

func (o *Orientation) Stable() bool {
	return o.imu.gyrCal.Stable
}

func (o *Orientation) InitialAngles() [3]float32 {
	return o.angles[0]
}

func (o *Orientation) Angles() [3]float32 {
	return o.angles[1]
}

func (o *Orientation) Offsets() (roll, pitch, yaw int32) {
	return o.imu.gyrCal.Offset[0], o.imu.gyrCal.Offset[1], o.imu.gyrCal.Offset[2]
}
