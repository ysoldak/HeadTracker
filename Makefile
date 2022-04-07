TARGET=nano-33-ble-s140v7
SIZE=full
DEBUG_OPT=1
SRC=./src

SOFTDEVICE_HEX=../bluetooth/s140_nrf52_7.3.0/s140_nrf52_7.3.0_softdevice.hex
ARDUINO_BOOTLOADER_BIN=~/Library/Arduino15/packages/arduino/hardware/mbed_nano/2.5.2/bootloaders/nano33ble/bootloader.bin

.PHONY: softdevice jlink-softdevice build build-for-debug flash jlink-flash nobt-flash debug jlink-debug restore

softdevice:
	openocd -f interface/cmsis-dap.cfg -f target/nrf52.cfg -c "transport select swd" -c "program $(SOFTDEVICE_HEX) verify reset exit"

jlink-softdevice:
	nrfjprog -f nrf52 --eraseall
	nrfjprog -f nrf52 --program $(SOFTDEVICE_HEX)

build:
	tinygo build -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=main -o ./build/build.hex $(SRC)

build-for-debug:
	tinygo build -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -o ./build/debug.elf $(SRC)

flash:
	tinygo flash -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=main -programmer=cmsis-dap $(SRC)

jlink-flash:
	tinygo flash -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=main $(SRC)

nobt-flash:
	tinygo flash -target=nano-33-ble -size=$(SIZE) -opt=z -print-allocs=main $(SRC)

debug: build-for-debug
	tinygo gdb -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=cmsis-dap $(SRC)

jlink-debug: build-for-debug
	tinygo gdb -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=jlink $(SRC)

restore:
	openocd -f interface/cmsis-dap.cfg -f target/nrf52.cfg -c "transport select swd" -c "program $(ARDUINO_BOOTLOADER_BIN) verify reset exit"

jlink-restore:
	nrfjprog -f nrf52 --eraseall
	nrfjprog -f nrf52 --program $(ARDUINO_BOOTLOADER_BIN)

# ---

build-xiao:
	tinygo build -target=xiao-ble -size=full -opt=z -print-allocs=main -o ./build/build.uf2 ./src

flash-xiao:
	tinygo flash -target=xiao-ble -size=full -opt=z -print-allocs=main -port=/dev/cu.usbmodem14201 -serial=uart ./src

