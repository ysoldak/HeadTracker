package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pca9685"
)

const (
	pcaAddr = 0x40
)

func main() {

	time.Sleep(2 * time.Second)

	err := machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.SDA0_PIN,
		SCL:       machine.SCL0_PIN,
	})
	handleError(err)

	d := pca9685.New(machine.I2C0, pcaAddr)
	d.SetInverting(0, true)
	err = d.IsConnected()
	handleError(err)

	err = d.Configure(pca9685.PWMConfig{Period: 1e9 / 50}) // 50Hz PWM = 20ms period
	handleError(err)

	micros := 1500 // 1500 microseconds = neutral position for most servos
	duty := (d.Top() * uint32(micros)) / 20000
	d.SetAll(duty)
	time.Sleep(5 * time.Second)

	// ticker := time.NewTicker(100 * time.Millisecond)
	// for range ticker.C {
	// para.Read()
	// println(para.Channels[2])
	// }

	para := NewParaTrainer()
	para.callbackConnected = func() {
		println("ParaTrainer connected")
	}
	para.callbackDisconnected = func() {
		println("ParaTrainer disconnected")
	}
	go para.Scan()

	go func() {
		for {
			para.Read()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	time.Sleep(2 * time.Second)

	step := 5
	delay := 100

	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {

		paraScaled := 1500 - (int16(para.Channels[2])-1500)*39/10
		println(micros, " = ", paraScaled, " diff ", int16(micros)-int16(paraScaled))

		if micros%100 == 0 {
			if delay > 0 {
				delay--
				continue
			}
			delay = 100
		}

		micros += step
		if micros > 2000 || micros < 1000 {
			micros -= step
			step = -step
			if step > 0 {
				step = 5
			} else {
				step = -100
			}
			micros += step
		}

		duty := (d.Top() * uint32(micros)) / 20000
		d.SetAll(duty)

		// println(para.Channels[2])
		// println("Set servo pulse width to", micros, "microseconds (duty cycle:", duty, ")")
	}

	// var value uint32
	// step := d.Top() / 5
	// for {
	// 	for value = 0; value <= d.Top(); value += step {
	// 		d.SetAll(value)
	// 		dc := 100 * value / d.Top()
	// 		println("set dc @", dc, "%")
	// 		time.Sleep(800 * time.Millisecond)
	// 	}
	// }
}

func handleError(err error) {
	if err == nil {
		return
	}
	for {
		println("Error:", err.Error())
		time.Sleep(1 * time.Second)
	}
}
