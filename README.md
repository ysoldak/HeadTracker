# HeadTracker
Head Tracker based on [Arduino Nano 33 BLE](https://store.arduino.cc/arduino-nano-33-ble) that connects to [OpenTX](https://github.com/opentx/opentx) radio via Bluetooth.

## Tested radios
- FrSky X-Lite Pro (OpenTX version 2.3.11)

## Dependencies
- [Custom version of ArduinoBLE library](https://github.com/ysoldak/ArduinoBLE/tree/cccd_hack)
  [A little hack](https://github.com/ysoldak/ArduinoBLE/compare/master...ysoldak:cccd_hack) to enable notification sending (CCCD descriptor on FFF6 characteristic)
  Please download "cccd_hack" branch of linked fork repository and replace original ArduinoBLE library with it

## Test Bluetooth connectivity
Steps:
- Install **custom version of ArduinoBLE** library (see above)
- Flash your Arduino Nano 33 BLE board with `test-bluetooth.ino` sketch
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
