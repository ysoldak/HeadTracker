_This information is only actual for Arduino Nano 33 BLE boards, since Seeeduino XIAO BLE ships with UF2 bootloader pre-flashed._

## Flash UF2 bootloader on Nano 33 BLE board

You'll need a JLink \[EDU\] or any DAP-compatible [debugger](https://tinygo.org/docs/guides/debugging/#connecting-a-debug-probe) to flash [Adafruit's nRF52 bootloader](https://github.com/adafruit/Adafruit_nRF52_Bootloader) with soft device on Nano 33 BLE board.

Connect your debugger to the board and execute: `make flash-uf2-bootloader`.

_Tip: by default make uses JLink debugger, but there are `-dap` versions of above commands too._

---

## Optionally use S140 v7.3.0 soft device
_This step is not necessary and is provided for informational purpose only_

Adafruit pre-compiles UF2 bootloader with soft device version 6.1.1 for Nano 33 BLE, while newer (better?) soft device version 7.3.0 is available.

To compile UF2 bootloader with 7.3.0 soft device:
- Clone [Adafruit's nRF52 bootloader](https://github.com/adafruit/Adafruit_nRF52_Bootloader) repository
- Run `make BOARD=arduino_nano_33_ble SD_VERSION=7.3.0 all` to compile bootloader with 7.3.0 soft device
- Connect your board to JLink debugger
- Run `nrfjprog -f nrf52 --eraseall` to erase chip
- Run `nrfjprog -f nrf52 --program <path to arduino_nano_33_ble_bootloader-0.7.0_s140_7.3.0.hex>` to flash you board with bootloader and soft device
- Run `make TARGET=nano-33-ble-s140v7-uf2 flash` to flash HeadTracker to Nano 33 BLE board with 7.3.0 soft device
