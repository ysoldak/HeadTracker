//go:build nano_33_ble

package orientation

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
	return &IMU{
		gyrCal: &GyrCal{},
	}
}

func (imu *IMU) Configure() error {
	// Configure I2C
	err := machine.I2C1.Configure(machine.I2CConfig{
		Frequency: 100 * machine.KHz,
		SDA:       machine.SDA1_PIN,
		SCL:       machine.SCL1_PIN,
	})
	if err != nil {
		return err
	}

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Configure IMU
	imu.device = lsm9ds1.New(machine.I2C1)
	err = imu.device.Configure(lsm9ds1.Configuration{
		AccelRange:      lsm9ds1.ACCEL_4G,
		AccelSampleRate: lsm9ds1.ACCEL_SR_238,
		GyroRange:       lsm9ds1.GYRO_500DPS,
		GyroSampleRate:  lsm9ds1.GYRO_SR_238,
		MagRange:        lsm9ds1.MAG_4G,
		MagSampleRate:   lsm9ds1.MAG_SR_80,
	})

	return err
}

func (imu *IMU) Read() (gx, gy, gz, ax, ay, az float64, err error) {
	gxi, gyi, gzi, err := imu.device.ReadRotation()
	if err != nil {
		println(err)
		return 0, 0, 0, 0, 0, 0, err
	}
	axi, ayi, azi, err := imu.device.ReadAcceleration()
	if err != nil {
		println(err)
		return 0, 0, 0, 0, 0, 0, err
	}

	imu.gyrCal.Apply(gxi, gyi, gzi)
	gxi, gyi, gzi = imu.gyrCal.Get(gxi, gyi, gzi)

	gx, gy, gz = float64(-gxi)/1000000, float64(gyi)/1000000, float64(gzi)/1000000
	ax, ay, az = float64(-axi)/1000000, float64(ayi)/1000000, float64(azi)/1000000
	return
}

func (imu *IMU) ReadTap() (tap bool) {
	return false // TODO implemented tap detection on Nano 33 BLE
}
