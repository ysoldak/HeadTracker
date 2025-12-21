package orientation

import (
	"math"
	"time"

	mgl "github.com/go-gl/mathgl/mgl64"
	"github.com/tracktum/go-ahrs"
)

const radToDeg = 180 / math.Pi // 57.29578
const degToRad = 1 / radToDeg  // 0.0174533

// Rule of thumb: increasing beta leads to (a) faster bias corrections, (b) higher sensitiveness to lateral accelerations.
// https://stackoverflow.com/questions/47589230/what-is-the-best-beta-value-in-madgwick-filter
const madgwickBeta = 0.025

type Orientation struct {
	imu     *IMU
	fusion  ahrs.Madgwick
	offset  mgl.Quat
	current mgl.Quat
}

func New(imu *IMU) *Orientation {
	return &Orientation{
		imu:    imu,
		offset: mgl.QuatIdent(),
	}
}

func (o *Orientation) Configure(period time.Duration) error {
	err := o.imu.Configure()
	if err != nil {
		return err
	}
	o.fusion = ahrs.NewMadgwick(madgwickBeta, float64(time.Second/period))
	return nil
}

// Reset orientation for sensor fusion algoritm
// - aligns current gravitation vector with Z axis
// - resets fusion quaternion
func (o *Orientation) Reset() {
	_, _, _, ax, ay, az, err := o.imu.Read()
	if err != nil {
		println(err.Error())
		return
	}
	start := mgl.Vec3{ax, ay, az}
	dest := mgl.Vec3{0, 0, 1}
	o.offset = mgl.QuatBetweenVectors(start, dest)
	o.fusion.Quaternions = [4]float64{1, 0, 0, 0}
}

// Calibrate gyroscope
func (o *Orientation) Calibrate() (corrections [3]int32) {
	_, _, _, _, _, _, err := o.imu.Read()
	if err != nil {
		println(err.Error())
	}
	return o.imu.gyrCal.correctionLast
}

// Update orientation
func (o *Orientation) Update() {
	// read raw data
	gx, gy, gz, ax, ay, az, err := o.imu.Read()
	if err != nil {
		println(err.Error())
		return
	}
	// rotate raw vectors to original offset
	a := o.offset.Rotate(mgl.Vec3{ax, ay, az})
	g := o.offset.Rotate(mgl.Vec3{gx, gy, gz})
	// apply fusion
	q := o.fusion.Update6D(
		g[0]*degToRad, g[1]*degToRad, g[2]*degToRad,
		a[0], a[1], a[2],
	)
	o.current.W = q[0]
	o.current.V = mgl.Vec3{q[1], q[2], q[3]}
}

// Angles in radians
func (o *Orientation) Angles() (angles [3]float64) {
	q := o.current
	angles[0] = math.Atan2(2*(q.W*q.V.X()+q.V.Y()*q.V.Z()), 1-2*(q.V.X()*q.V.X()+q.V.Y()*q.V.Y()))
	angles[1] = math.Asin(2 * (q.W*q.V.Y() - q.V.X()*q.V.Z()))
	angles[2] = math.Atan2(2*(q.V.X()*q.V.Y()+q.W*q.V.Z()), 1-2*(q.V.Y()*q.V.Y()+q.V.Z()*q.V.Z()))
	return
}

// Stable state indicates gyroscope calibration is good
func (o *Orientation) Stable() bool {
	return o.imu.gyrCal.Stable
}

func (o *Orientation) SetStable(stable bool) {
	o.imu.gyrCal.Stable = stable
}

func (o *Orientation) Offsets() (offsets [3]int32) {
	return o.imu.gyrCal.Offset
}

func (o *Orientation) SetOffsets(offsets [3]int32) {
	o.imu.gyrCal.Offset = offsets
	if offsets[0] == 0 && offsets[1] == 0 && offsets[2] == 0 {
		return
	}
	// calibration shall skip aggressive first force adjustments when data is non-zero
	o.imu.gyrCal.countAdjust[0] = gyrCalForceAdjustment + 1
	o.imu.gyrCal.countAdjust[1] = gyrCalForceAdjustment + 1
	o.imu.gyrCal.countAdjust[2] = gyrCalForceAdjustment + 1
}
