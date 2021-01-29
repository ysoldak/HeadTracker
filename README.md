# Head Tracker
Head Tracker that runs on [Arduino Nano 33 BLE](https://store.arduino.cc/arduino-nano-33-ble) and connects to [OpenTX](https://github.com/opentx/opentx) radios via Bluetooth.

**Arduino Nano 33 BLE** board if perfect for head tracker project, since it has both **9DOF IMU** for orientation and **Bluetooth** for connectivity.

Before you begin, install patched versions of Bluetooth and IMU libraries (see below).

Upload `main` sketch, pair with radio, configure **TR7** and **TR8** inputs to be sent to your model for **pan** and **tilt** servos.

Good idea is to assign an override switch for pan and tilt channels, so you can always center your camera if something goes wrong with the head tracker. This is also convenient when you about to put your goggles on and do not want camera servos going mad.

Default maximum angles: **45deg** each side, can be adjusted if your servos capable of more.

Do not forget to calibrate magnetometer.

## Tested radios
- FrSky X-Lite Pro (OpenTX version 2.3.11)

## Dependencies
- [ArduinoBLE](https://github.com/ysoldak/ArduinoBLE/tree/cccd_hack), custom version ([diff](https://github.com/ysoldak/ArduinoBLE/compare/master...ysoldak:cccd_hack))  
  A little hack to enable notification sending (CCCD descriptor on FFF6 characteristic)  
- [Arduino_LSM9DS1](https://github.com/ysoldak/Arduino_LSM9DS1/tree/head_tracker_settings), custom version ([diff](https://github.com/ysoldak/Arduino_LSM9DS1/compare/master...ysoldak:head_tracker_settings))  
  Tweaked rates and disabled Gyro

Please download linked branches of above repositories and replace original libraries with them.

## Calibrate Magnetometer
Run `calibration/calibration.ino` sketch, connect with serial console and follow instructions.
When happy with results, copy calibration data and paste at respective place into `main/imu.cpp`.

Note: Accelermeter usually does not require calibration.

## Test Bluetooth connectivity
Steps:
- Install **custom version of ArduinoBLE** library (see above)
- Flash your Arduino Nano 33 BLE board with `test-bluetooth/test-bluetooth.ino` sketch
- Connect to the board with **Serial console** and make note of the board **address** (something like: `7b:f5:1e:35:de:94`)
- In your radio, select Trainer mode **"Master/BT"**, wait a bit and click "[Discover]"
- Search for your Arduino board by address you noted earlier and **Connect** to it
- **Built-in led** on the board shall turn **on** and in serial console you shall see "Connected to central" and "Subscribed"
- Test sketch sends constant channel values: all channels shall be on min (~1000), and only channel 3 (throttle) is on max (~2000)
- Do not forget to configure **Trainer function** in your radio either on "Special Functions" screen of your model or on "Global Functions" of your radio setup.

Notes:
- Sometimes it would not connect at the first attempt (even when radio thinks it's connected, led stays off), keep trying.
- Good results were achieved by switching trainer mode to "Slave/BT" briefly and then back to "Master/BT".
- Once connection is established it stays pretty stable. You may even try power-cycle your board, in this case radio will warn you "trainer signal lost" followed by "trainer signal recovered".
- Serial console connection is not needed for successfull test and connection, you really need it only once -- to learn your board address.
- Connection status is indicated with the built-in led.

## Related links
- [DIY-Head-Tracker](https://github.com/kniuk/DIY-Head-Tracker)  
  Original DIY head tracker for Arduino Nano with separate IMU board and PPM over cable
- [Arduino_LSM9DS1 V2](https://github.com/FemmeVerbeek/Arduino_LSM9DS1)  
  Lots of information on how to work with magnetic field
- [Bluetooth low energy Characteristics, a beginner's tutorial](https://devzone.nordicsemi.com/nordic/short-range-guides/b/bluetooth-low-energy/posts/ble-characteristics-a-beginners-tutorial)
