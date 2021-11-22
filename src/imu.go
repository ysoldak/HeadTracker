package main

import (
	"machine"
	"math"
	"time"

	"github.com/tracktum/go-ahrs"
	"github.com/ysoldak/magcal"
	"tinygo.org/x/drivers/lsm9ds1"
)

var imu *lsm9ds1.Device
var gyrCal *GyrCal
var magCal *magcal.MagCal
var fusion ahrs.Mahony

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
	magCal = magcal.NewDefault()
	magCal.Start()

	// Fusion
	fusion = ahrs.NewDefaultMahony(float64(time.Second / PERIOD))
}

func imuWarmup() {
	ax, ay, az, _ := imu.ReadAcceleration()
	gx, gy, gz, _ := imu.ReadRotation()
	mx, my, mz, _ := imu.ReadMagneticField()

	mxf, myf, mzf := float32(mx)/50000, float32(my)/50000, float32(mz)/50000
	mxf, myf, mzf = magCal.Apply(mxf, myf, mzf)

	// TODO fusion init
	fusion.Update9D(
		float64(gx), float64(gy), float64(gz),
		float64(ax), float64(ay), float64(az),
		float64(mxf), float64(myf), float64(mzf),
	)
}

func imuAngles() (pitch, roll, yaw float64) {
	ax, ay, az, _ := imu.ReadAcceleration()
	gx, gy, gz, _ := imu.ReadRotation()
	mx, my, mz, _ := imu.ReadMagneticField()

	mxf, myf, mzf := float32(mx)/50000, float32(my)/50000, float32(mz)/50000
	mxf, myf, mzf = magCal.Apply(mxf, myf, mzf)

	q := fusion.Update9D(
		float64(gx), float64(gy), float64(gz),
		float64(ax), float64(ay), float64(az),
		float64(mxf), float64(myf), float64(mzf),
	)
	roll, pitch, yaw = qToAngles(q)
	return
}

func qToAngles(q [4]float64) (roll, pitch, yaw float64) {
	roll = math.Atan2(q[0]*q[1]+q[2]*q[3], 0.5-q[1]*q[1]-q[2]*q[2])
	pitch = math.Asin(-2.0 * (q[1]*q[3] - q[0]*q[2]))
	yaw = math.Atan2(q[1]*q[2]+q[0]*q[3], 0.5-q[2]*q[2]-q[3]*q[3])
	return
}
