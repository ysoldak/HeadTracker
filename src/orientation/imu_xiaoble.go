//go:build xiao_ble
// +build xiao_ble

package orientation

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm6ds3"
)

type IMU struct {
	device *lsm6ds3.Device
	gyrCal *GyrCal
}

func NewIMU() *IMU {
	return &IMU{
		gyrCal: &GyrCal{},
	}
}

func (imu *IMU) Configure() {
	// Configure I2C
	machine.I2C1.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_100KHZ,
		SDA:       machine.SDA1_PIN,
		SCL:       machine.SCL1_PIN,
	})

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Configure IMU
	imu.device = lsm6ds3.New(machine.I2C1)
	imu.device.Configure(lsm6ds3.Configuration{
		AccelRange:      lsm6ds3.ACCEL_4G,
		AccelSampleRate: lsm6ds3.ACCEL_SR_104,
		GyroRange:       lsm6ds3.GYRO_250DPS,
		GyroSampleRate:  lsm6ds3.GYRO_SR_104,
	})

}

func (imu *IMU) Read(calibrate bool) (gx, gy, gz, ax, ay, az float64, err error) {
	gxi, gyi, gzi, err := imu.device.ReadRotation()
	for err != nil {
		println(err)
		return 0, 0, 0, 0, 0, 0, err
	}
	axi, ayi, azi, err := imu.device.ReadAcceleration()
	for err != nil {
		println(err)
		return 0, 0, 0, 0, 0, 0, err
	}

	if calibrate {
		imu.gyrCal.apply(gxi, gyi, gzi)
	}
	gxi, gyi, gzi = imu.gyrCal.get(gxi, gyi, gzi)

	gx, gy, gz = float64(gxi)/1000000, float64(-gyi)/1000000, float64(-gzi)/1000000
	ax, ay, az = float64(-axi)/1000000, float64(ayi)/1000000, float64(azi)/1000000
	return
}
