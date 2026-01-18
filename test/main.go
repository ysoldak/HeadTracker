package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pca9685"
)

const CHANNEL_INDEX = 2
const INVERSE = true

const (

	// base scenario (approximation of the human behaviour)
	// step fast one side, even faster other side and wait a bit at 100s
	// example results: 6000 : 1810 | 1579 = 1615 diff  -36     -52 @ 1669 / 50 @ 2762
	// STEP_POSITIVE = 10
	// STEP_NEGATIVE = -100
	// WAIT_AT       = 100

	// sanity check for the base scenario, shall give similar results
	// step faster one side, normal other side and wait a bit at 100s
	// STEP_POSITIVE = 100
	// STEP_NEGATIVE = -5
	// WAIT_AT = 100

	// fast sweep
	// step normal one side, very fast other side and wait a bit at 1000s
	// example results: 6000 : 2000 | 1628 = 1571 diff  57     -37 @ 733 / 189 @ 5125
	STEP_POSITIVE = 5
	STEP_NEGATIVE = -100
	WAIT_AT       = 1000

	// slow sweep
	// step slow one side, very fast other side and wait a bit at 1000s
	// STEP_POSITIVE = 1
	// STEP_NEGATIVE = -100
	// WAIT_AT       = 1000

	DELAY  = 100
	PERIOD = 100
)

const (
	pcaAddr = 0x40
)

func main() {

	time.Sleep(2 * time.Second)

	// ---------------------------------------------
	// Setup PCA9685 for servo control
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

	// Setup head tracker connection
	para := NewParaTrainer()
	para.callbackConnected = func() {
		println("ParaTrainer connected")
	}
	para.callbackDisconnected = func() {
		println("ParaTrainer disconnected")
	}
	go para.Start()

	time.Sleep(2 * time.Second)

	// ---------------------------------------------
	// Main loop moves servo and reads head tracker

	step := STEP_POSITIVE
	delay := DELAY
	iteration := 0
	iterationMax := 10 * 60 * 10 // 10 minutes

	diffMin := int16(0)
	diffMinIter := 0
	diffMax := int16(0)
	diffMaxIter := 0

	ticker := time.NewTicker(PERIOD * time.Millisecond)
	for range ticker.C {

		iteration++
		if iteration > iterationMax {
			break
		}

		// Following is just for outputs
		// 1. inverse servo command if needed
		outMicros := uint16(micros)
		if INVERSE {
			outMicros = 3000 - outMicros
		}
		// 2. scale servo command to match expected values from the head tracker, as HT uses [988:2012] range for 360 degrees
		outMicrosScaled := 1500 + (int16(outMicros)-1500)*10/39 // expected value
		// 3. value received from head tracker
		outPara := para.Channels[CHANNEL_INDEX]
		// 4. calculate difference
		diff := int16(outMicrosScaled) - int16(outPara)
		if diff < diffMin {
			diffMin = diff
			diffMinIter = iteration
		}
		if diff > diffMax {
			diffMax = diff
			diffMaxIter = iteration
		}
		// 5. print results
		println(iteration, ":", outMicros, "|", outMicrosScaled, "=", outPara, "diff ", diff, "   ", diffMin, "@", diffMinIter, "/", diffMax, "@", diffMaxIter)

		// pause at 100s or 1000s or whatever WAIT_AT is set to
		if micros%WAIT_AT == 0 {
			if delay > 0 {
				delay--
				continue // this short-circuits the loop and skips update servo position part
			}
			delay = DELAY
		}

		// update servo position
		micros += step
		if micros > 2000 || micros < 1000 {
			micros -= step
			step = -step
			if step > 0 {
				step = STEP_POSITIVE
			} else {
				step = STEP_NEGATIVE
			}
			micros += step
		}

		// set servo position
		duty := (d.Top() * uint32(micros)) / 20000
		d.SetAll(duty)

	}

	// center servo in the end
	duty = (d.Top() * uint32(1500)) / 20000
	d.SetAll(duty)

	// wait forever
	for {
		time.Sleep(1 * time.Second)
	}

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
