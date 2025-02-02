//go:build xiao_ble

package orientation

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lsm6ds3tr"
)

const (
	TAP_SRC     = 0x1C
	TAP_CFG     = 0x58
	TAP_THS_6D  = 0x59
	INT_DUR2    = 0x5A
	WAKE_UP_THS = 0x5B
	MD1_CFG     = 0x5E
)

type IMU struct {
	device *lsm6ds3tr.Device
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
		Frequency: 100 * machine.KHz,
		SDA:       machine.SDA1_PIN,
		SCL:       machine.SCL1_PIN,
	})

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Configure IMU
	imu.device = lsm6ds3tr.New(machine.I2C1)
	imu.device.Configure(lsm6ds3tr.Configuration{
		AccelRange:      lsm6ds3tr.ACCEL_4G,     // 4g
		AccelSampleRate: lsm6ds3tr.ACCEL_SR_208, // every ~4.8ms
		GyroRange:       lsm6ds3tr.GYRO_500DPS,  // 500 deg/s
		GyroSampleRate:  lsm6ds3tr.GYRO_SR_208,  // every ~4.8ms
	})

	tapConfig := map[byte]byte{
		TAP_CFG:     0x8F, // interrupts enable + tap all axes + latch (saves the state of the interrupt until register is read)
		TAP_THS_6D:  0x02, // tap threshold
		INT_DUR2:    0x20, // tap sensing params: duration = 16*([7:4]+1)*4.8ms, quiet = 2*([3:2]+1)*4.8ms, shock = 4*([1:0]+1)*4.8ms => 0x20 = 230.4ms, 9.6ms, 19.2ms
		WAKE_UP_THS: 0x80, // enable double tap events
		MD1_CFG:     0x08, // route double tap events to INT1 (requited for the latch to work)
	}

	for reg, val := range tapConfig {
		machine.I2C1.WriteRegister(uint8(imu.device.Address), reg, []byte{val})
	}

}

func (imu *IMU) Read() (gx, gy, gz, ax, ay, az float32, err error) {
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

	gx, gy, gz = float32(gxi)/1000000, float32(-gyi)/1000000, float32(-gzi)/1000000
	ax, ay, az = float32(-axi)/1000000, float32(ayi)/1000000, float32(azi)/1000000
	return
}

func (imu *IMU) ReadTap() (tap bool) {
	data := []byte{0x00}
	machine.I2C1.ReadRegister(uint8(imu.device.Address), TAP_SRC, data)
	return data[0]&0x10 != 0
}
