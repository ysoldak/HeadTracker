package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm9ds1"
)

type IMU struct {
	device *lsm9ds1.Device
	gyrCal *GyrCal
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

}

func (imu *IMU) Read() (gx, gy, gz, ax, ay, az float64) {
	gxi, gyi, gzi, _ := imu.device.ReadRotation()
	axi, ayi, azi, _ := imu.device.ReadAcceleration()

	gxi, gyi, gzi = imu.gyrCal.apply(gxi, gyi, gzi)

	gx, gy, gz = float64(-gxi)/1000000, float64(gyi)/1000000, float64(gzi)/1000000
	ax, ay, az = float64(-axi)/1000000, float64(ayi)/1000000, float64(azi)/1000000
	return
}
