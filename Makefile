TARGET=nano-33-ble-s140v7
SIZE=short
DEBUG_OPT=1
SRC=./src/go

.PHONY: softdevice jlink-softdevice flash jlink-flash build-for-debug debug jlink-debug

softdevice:
	openocd -f interface/cmsis-dap.cfg -f target/nrf52.cfg -c "transport select swd" -c "program ../bluetooth/s140_nrf52_7.3.0/s140_nrf52_7.3.0_softdevice.hex verify reset exit"

jlink-softdevice:
	nrfjprog -f nrf52 --eraseall
	nrfjprog -f nrf52 --program ../bluetooth/s140_nrf52_7.3.0/s140_nrf52_7.3.0_softdevice.hex

flash:
	tinygo flash -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=main -programmer=cmsis-dap $(SRC)

jlink-flash:
	tinygo flash -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=main $(SRC)

build-for-debug:
	tinygo build -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -o ./build/debug.elf $(SRC)

debug: build-for-debug
	tinygo gdb -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=cmsis-dap $(SRC)

jlink-debug: build-for-debug
	tinygo gdb -target=$(TARGET) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=jlink $(SRC)
