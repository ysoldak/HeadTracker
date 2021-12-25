package main

import (
	"fmt"
	"machine"
	"math"
	"time"

	"github.com/tracktum/go-ahrs"
	"github.com/ysoldak/magcal"
	"tinygo.org/x/drivers/lsm9ds1"
)

const radToDeg = 180 / math.Pi // 57.29578
const degToRad = 1 / radToDeg  // 0.0174533

var imu *lsm9ds1.Device
var gyrCal *GyrCal
var magCal *magcal.MagCal

// var fusion ahrs.Mahony
var fusion ahrs.Madgwick

var startAngles = [3]float32{0, 0, 0}

func imuSetup() {
	// Board setup
	machine.I2C1.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SDA:       machine.SDA1_PIN,
		SCL:       machine.SCL1_PIN,
	})
	time.Sleep(1 * time.Second)

	// LSM9DS1 setup
	imu = lsm9ds1.New(machine.I2C1)
	imu.Configure(lsm9ds1.Configuration{
		AccelRange:      lsm9ds1.ACCEL_4G,
		AccelSampleRate: lsm9ds1.ACCEL_SR_119,
		GyroRange:       lsm9ds1.GYRO_250DPS,
		GyroSampleRate:  lsm9ds1.GYRO_SR_119,
		MagRange:        lsm9ds1.MAG_4G,
		MagSampleRate:   lsm9ds1.MAG_SR_80,
	})

	gyrCal = &GyrCal{}

	// Magnetometer runtime calibration
	magCalConfig := magcal.DefaultConfiguration()
	magCalConfig.Throttle = PERIOD
	magCalState := magcal.NewState([]float32{
		-0.27, -0.16, +0.20,
		+1.13, +0.06, -0.03,
		+0.05, +1.21, -0.03,
		-0.01, -0.02, +1.21,
	})
	// magCalState := magcal.DefaultState()
	magCal = magcal.New(magCalState, magCalConfig)
	// magCal.Start()

	// Fusion
	// fusion = ahrs.NewDefaultMahony(float64(time.Second / PERIOD))
	// fusion = ahrs.NewMadgwick(0.2, float64(time.Second/PERIOD))
	fusion = ahrs.NewMadgwick(0.5, 20)
}

func imuWork(output chan ([3]float32)) {
	warmup := time.Now().Add(5 * time.Second)
	// warmup := time.Now().Add(5 * time.Hour)
	i := 0
	for {
		i++
		now := time.Now()
		if now.Before(warmup) {
			r, p, y := imuWarmup()
			startAngles[0] = r
			startAngles[1] = p
			startAngles[2] = y
			// if i%100 == 0 {
			// 	StateDump(magcal.State)
			// }
			// qToAngles(fusion.Quaternions)
			// fmt.Printf("%+3.3f %+3.3f %+3.3f\r\n", roll, pitch, yaw)
			// TODO remember start angles
		} else {
			// panic("AAA!!!")
			r, p, y := imuAngles()
			// roll, pitch, yaw := qToAngles(fusion.Quaternions)
			// fmt.Printf("%+000.3f %+000.3f %+000.3f\r\n", r, p, y)
			arr := [3]float32{
				imuAngleDiff(startAngles[0], r),
				imuAngleDiff(startAngles[1], p),
				imuAngleDiff(startAngles[2], y),
			}
			select {
			case output <- arr:
			default:
				println("imu: output channel full, dropping measurement")
			}
		}
		sleep := PERIOD - time.Since(now)
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

func imuWarmup() (roll, pitch, yaw float32) {
	gx, gy, gz, _ := imu.ReadRotation()
	ax, ay, az, _ := imu.ReadAcceleration()
	mx, my, mz, _ := imu.ReadMagneticField()

	gx, gy, gz = gyrCal.apply(gx, gy, gz)

	mxf, myf, mzf := float32(mx)/50000, float32(my)/50000, float32(mz)/50000
	mxf, myf, mzf = magCal.Apply(mxf, myf, mzf)

	fusion.Quaternions = orientationToQ(
		float64(-ax)/1000000, float64(ay)/1000000, float64(az)/1000000,
		float64(mxf), float64(myf), float64(mzf),
	)

	roll, pitch, yaw = qToAngles(fusion.Quaternions)
	return
}

func imuAngles() (roll, pitch, yaw float32) {
	gx, gy, gz, _ := imu.ReadRotation()
	ax, ay, az, _ := imu.ReadAcceleration()
	mx, my, mz, _ := imu.ReadMagneticField()

	mxf, myf, mzf := float32(mx)/50000, float32(my)/50000, float32(mz)/50000
	mxf, myf, mzf = magCal.Apply(mxf, myf, mzf)

	gx, gy, gz = gyrCal.apply(gx, gy, gz)

	q := fusion.Update9D(
		float64(-gx)/1000000*degToRad, float64(gy)/1000000*degToRad, float64(gz)/1000000*degToRad,
		float64(-ax)/1000000, float64(ay)/1000000, float64(az)/1000000,
		float64(mxf), float64(myf), float64(mzf),
	)
	// q := fusion.Update6D(
	// 	float64(-gx)/1000000, float64(gy)/1000000, float64(gz)/1000000,
	// 	float64(-ax)/1000000, float64(ay)/1000000, float64(az)/1000000,
	// )
	roll, pitch, yaw = qToAngles(q)
	return
}

// angles are in degrees
func qToAngles(q [4]float64) (roll, pitch, yaw float32) {
	roll = float32(math.Atan2(q[0]*q[1]+q[2]*q[3], 0.5-q[1]*q[1]-q[2]*q[2]) * radToDeg)
	pitch = float32(math.Asin(-2.0*(q[1]*q[3]-q[0]*q[2])) * radToDeg)
	yaw = float32(math.Atan2(q[1]*q[2]+q[0]*q[3], 0.5-q[2]*q[2]-q[3]*q[3]) * radToDeg)
	return
}

func imuAngleDiff(ref, now float32) float32 {
	diff := now - ref
	if diff < -180 {
		diff += 360
	}
	if diff > 180 {
		diff -= 360
	}
	return diff
}

// -- Helper --

func StateDump(data []float32) {
	println("===")
	for i := range data {
		fmt.Printf("%5.2f, ", data[i])
		if i > 0 && (i+1)%3 == 0 {
			println()
		}
	}
	println("===")
}

// -- Quaternion initialisation --

// Initialize quaternion from current orientation (sensors)
// Finds North, then aligns North with (1, 0, 0) and Gravity with (0, 0, 1)
func orientationToQ(ax, ay, az, mx, my, mz float64) (q [4]float64) {

	// Reset quaternion, we are searching from scratch
	q[0], q[1], q[2], q[3] = 1, 0, 0, 0

	// Find North
	wx, wy, wz := cross(ax, ay, az, mx, my, mz)
	nx, ny, nz := cross(wx, wy, wz, ax, ay, az)

	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", ax, ay, az)
	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", mx, my, mz)
	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", nx, ny, nz)

	// Find rotation of (nx, ny, nz) to align with (1, 0, 0)
	q = align(nx, ny, nz, 1, 0, 0, q)

	// Rotate (ax, ay, az) same amount
	ax, ay, az = rotate(ax, ay, az, q)

	// Find next rotation of (ax, ay, az) to align with (0, 0, 1)
	q = align(ax, ay, az, 0, 0, 1, q)
	return
}

// align two vectors and return respective rotation quaternion
func align(ax, ay, az, bx, by, bz float64, u [4]float64) (q [4]float64) {
	// float va, vx, vy, vz; // rotation angle and vector
	vx, vy, vz := cross(ax, ay, az, bx, by, bz)
	ax, ay, az = norm(ax, ay, az)
	bx, by, bz = norm(bx, by, bz)
	vx, vy, vz = norm(vx, vy, vz)
	va := math.Acos(dot(ax, ay, az, bx, by, bz))
	a2 := math.Cos(va / 2)
	b2 := vx * math.Sin(va/2)
	c2 := vy * math.Sin(va/2)
	d2 := vz * math.Sin(va/2)
	v := [4]float64{a2, b2, c2, d2}
	q = combine(u, v)
	return
}

// combine two rotations and return result rotation
// see https://en.wikipedia.org/wiki/Euler–Rodrigues_formula#Composition_of_rotations
func combine(u, v [4]float64) (q [4]float64) {
	a1, b1, c1, d1 := u[0], u[1], u[2], u[3]
	a2, b2, c2, d2 := v[0], v[1], v[2], v[3]
	q[0] = a1*a2 - b1*b2 - c1*c2 - d1*d2
	q[1] = a1*b2 + b1*a2 - c1*d2 + d1*c2
	q[2] = a1*c2 + c1*a2 - d1*b2 + b1*d2
	q[3] = a1*d2 + d1*a2 - b1*c2 + c1*b2
	return
}

// rotate given vector as defined by quaternion
// see https://en.wikipedia.org/wiki/Euler–Rodrigues_formula#Vector_formulation
func rotate(ax, ay, az float64, q [4]float64) (bx, by, bz float64) {
	r1x, r1y, r1z := cross(q[1], q[2], q[3], ax, ay, az)
	r2x, r2y, r2z := cross(q[1], q[2], q[3], r1x, r1y, r1z)
	bx = ax + 2*q[0]*r1x + 2*r2x
	by = ay + 2*q[0]*r1y + 2*r2y
	bz = az + 2*q[0]*r1z + 2*r2z
	return
}

// cross product of two vectors
func cross(ax, ay, az, bx, by, bz float64) (cx, cy, cz float64) {
	cx = ay*bz - az*by
	cy = az*bx - ax*bz
	cz = ax*by - ay*bx
	return
}

// dot product of two vectors
func dot(ax, ay, az, bx, by, bz float64) float64 {
	return ax*bx + ay*by + az*bz
}

// norm of a vector
func norm(ax, ay, az float64) (bx, by, bz float64) {
	len := math.Sqrt(dot(ax, ay, az, ax, ay, az))
	bx = ax / len
	by = ay / len
	bz = az / len
	return
}
