package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/ysoldak/magcal"
	"tinygo.org/x/drivers/lsm9ds1"
)

type IMU struct {
	device *lsm9ds1.Device
	gyrCal *GyrCal
	magCal *magcal.MagCal
}

func NewIMU() *IMU {
	return &IMU{}
}

func (imu *IMU) Configure() {
	// Board setup
	machine.I2C1.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SDA:       machine.SDA1_PIN,
		SCL:       machine.SCL1_PIN,
	})
	time.Sleep(1 * time.Second)

	// LSM9DS1 setup
	imu.device = lsm9ds1.New(machine.I2C1)
	imu.device.Configure(lsm9ds1.Configuration{
		AccelRange:      lsm9ds1.ACCEL_4G,
		AccelSampleRate: lsm9ds1.ACCEL_SR_119,
		GyroRange:       lsm9ds1.GYRO_250DPS,
		GyroSampleRate:  lsm9ds1.GYRO_SR_119,
		MagRange:        lsm9ds1.MAG_4G,
		MagSampleRate:   lsm9ds1.MAG_SR_80,
	})

	// Gyroscope runtime calibration
	imu.gyrCal = &GyrCal{}

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
	imu.magCal = magcal.New(magCalState, magCalConfig)
	// magCal.Start()

}

func (imu *IMU) Read() (gx, gy, gz, ax, ay, az, mx, my, mz float64) {
	gxi, gyi, gzi, _ := imu.device.ReadRotation()
	axi, ayi, azi, _ := imu.device.ReadAcceleration()
	mxi, myi, mzi, _ := imu.device.ReadMagneticField()

	gxi, gyi, gzi = imu.gyrCal.apply(gxi, gyi, gzi)

	mxf, myf, mzf := float32(mxi)/50000, float32(myi)/50000, float32(mzi)/50000
	mxf, myf, mzf = imu.magCal.Apply(mxf, myf, mzf)

	gx, gy, gz = float64(-gxi)/1000000, float64(gyi)/1000000, float64(gzi)/1000000
	ax, ay, az = float64(-axi)/1000000, float64(ayi)/1000000, float64(azi)/1000000
	mx, my, mz = float64(mxf), float64(myf), float64(mzf)
	return
}

// -- Helper --

func MagCalStateDump(data []float32) {
	println("===")
	for i := range data {
		fmt.Printf("%5.2f, ", data[i])
		if i > 0 && (i+1)%3 == 0 {
			println()
		}
	}
	println("===")
}
