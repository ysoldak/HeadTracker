# Arduino Nano 33 BLE Head Tracker
Head Tracker that runs on [Arduino Nano 33 BLE](https://store.arduino.cc/arduino-nano-33-ble) and connects to [OpenTX](https://github.com/opentx/opentx) radios via Bluetooth.

<table><tr><td>
<img src="case/CaseOnGoggles.jpg" title="Case mounted on the left side" style="float: left;"/>
</td><td>
<img src="case/CaseOpen.jpg" title="Case open, showing wiring" style="float: right;"/>
</td></tr></table>

**Arduino Nano 33 BLE** board if perfect for head tracker project, since it has both **9DOF IMU** for orientation and **Bluetooth** for connectivity.

Before you begin, install special versions of Bluetooth, IMU and Madgwick libraries ([see below](#dependencies)).

Next important step is IMU calibration ([see below](#calibration)).

Upload `main` sketch, pair with radio, configure **TR1**, **TR2** and **TR3** as inputs for camera control servos.

Board orientation defines what trainer channels correspond to pan, tilt and roll of the camera.

Assign an override switch for camera control channels, so you can always center your camera if needed.

Default maximum angles: **45deg** each side, can be adjusted in `config.h` if your servos capable of more.

## Tested radios
- FrSky X-Lite Pro (OpenTX version 2.3.11)

## Dependencies
- [ArduinoBLE](https://github.com/ysoldak/ArduinoBLE/tree/cccd_hack), custom version ([diff](https://github.com/ysoldak/ArduinoBLE/compare/master...ysoldak:cccd_hack))  
  A little hack to enable notification sending (CCCD descriptor on FFF6 characteristic)  
- [Arduino_LSM9DS1 V2](https://github.com/FemmeVerbeek/Arduino_LSM9DS1)  
  An improved version of stock IMU library, can configure sensitivity and update rates, also supports simple calibration
- [MadgwickAHRS](https://github.com/ysoldak/MadgwickAHRS/tree/set-methods), custom version ([diff](https://github.com/ysoldak/MadgwickAHRS/compare/master...ysoldak:set-methods))  
  Sensor fusion algorithm for drift-less and jitter-free heading, custom version adds quaternion initialisation.

Please [install libraries](https://learn.adafruit.com/adafruit-all-about-arduino-libraries-install-use) from linked branches of above repositories.

## Calibration
### Gyroscope
Gyroscope shall return zeros when board is stationary. No easy calibration yet.
Calibration workaround:
- Uncomment debug output in `main/imu.cpp`
- Switch to `imuDebugLoop` in `main/main.ino`
- Upload to the board, connect with serial monitor and leave the board stationary
- Locate and note gyroscope readings, put them into calibration matrix in `main/conifg.h`, gyroscope offset row
- Upload again and verify gyroscope reads near-zeros when board is stationary
- Revert changes (except `config.h`) to the code before you upload to the board again!

### Accelermeter
Usually does not require calibration.

### Magnetometer
Run `calibration/calibration.ino` sketch, connect with serial console and follow instructions.
When happy with results, copy calibration data and paste at respective place into `main/config.h`.


## Test Bluetooth connectivity
Steps:
- Install **custom version of ArduinoBLE** library (see above)
- Flash your Arduino Nano 33 BLE board with `test-bluetooth/test-bluetooth.ino` sketch
- Connect to the board with **Serial console** and make note of the board **address** (something like: `7b:f5:1e:35:de:94`)
- In your radio, select Trainer mode **"Master/BT"**, wait a bit and click "[Discover]"
- Search for your Arduino board by address you noted earlier and **Connect** to it
- **Built-in led** on the board shall turn **on** and in serial console you shall see "Connected to central"
- Test sketch sends constant channel values: all channels shall be on min (~1000), and only channel 3 (throttle) is on max (~2000)
- Do not forget to configure **Trainer function** in your radio either on "Special Functions" screen of your model or on "Global Functions" of your radio setup.

Notes:
- Sometimes it would not connect at the first attempt (even when radio thinks it's connected, led stays off), keep trying.
- Good results were achieved by switching trainer mode to "Slave/BT" briefly and then back to "Master/BT".
- Once connection is established it stays pretty stable. You may even power-cycle your board, in this case radio will warn you "trainer signal lost" followed by "trainer signal recovered". Switching radio off and on again after connection/pairing was successful will also result in successful re-connect, even after days off. Just be patient and wait some time for radio to boot it's bluetooth, discover head tracker and connect to it. Trying to force connection will not speedup the process, instead may put radio's bluetooth stack in some weird state and you will have to pair your radio and head tracker again.
- Serial console connection is not needed for successfull test and connection, you really need it only once -- to learn your board address.
- Connection status is indicated with the built-in led.

## Related links
- [DIY-Head-Tracker](https://github.com/kniuk/DIY-Head-Tracker)  
  Original DIY head tracker for Arduino Nano with separate IMU board and PPM over cable
- [RC HeadTracker by Cliff](https://github.com/dlktdr/HeadTracker)  
  Another, more advanced version of head tracker, also based on Arduino Nano 33 BLE [Sense] board.
- [Arduino_LSM9DS1 V2](https://github.com/FemmeVerbeek/Arduino_LSM9DS1)  
  Lots of information on how to work with magnetic field
- [Bluetooth Smart/BLE Crash Course](https://inductive-kickback.com/projects/bluetooth-low-energy/bluetooth-smartble-crash-course/)
- [Bluetooth low energy Characteristics, a beginner's tutorial](https://devzone.nordicsemi.com/nordic/short-range-guides/b/bluetooth-low-energy/posts/ble-characteristics-a-beginners-tutorial)
