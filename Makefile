SIZE   ?= full
TARGET ?= xiao-ble

ifneq ($(TARGET),nano-33-ble-s140v6-uf2)
FILE = ht_$(TARGET)_$(VERSION).uf2
else
FILE = ht_nano-33-ble_$(VERSION).uf2
endif

.PHONY: clean build flash monitor

# --- Go maintenance targets ---

bump-deps:
	go get -u github.com/go-gl/mathgl@latest
	go get -u github.com/tracktum/go-ahrs@latest
	go get -u tinygo.org/x/bluetooth@latest
	go get -u tinygo.org/x/drivers@latest
	go get -u tinygo.org/x/tinydraw@latest
	go get -u tinygo.org/x/tinyfont@latest
	go mod tidy

llvm:
	brew install llvm@20
	brew install lld@20
	brew link --force --overwrite llvm@20
	brew link --force --overwrite lld@20

# --- Common targets ---

VERSION ?= $(shell git describe --tags)
LD_FLAGS := -ldflags="-X 'main.Version=$(VERSION)'" # https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications

EXTRA_TAGS = nogopls # to include all source files, even those excluded for gopls

clean:
	@rm -rf build
	@rm -rf softdevice

softdevice:
	@mkdir -p softdevice
	@bluetooth_ver=$$(cat go.mod | grep bluetooth | sed "s/.* //"); \
	ln -s $(HOME)/go/pkg/mod/tinygo.org/x/bluetooth@$$bluetooth_ver/s140_nrf52_6.1.1 ./softdevice/; \
	ln -s $(HOME)/go/pkg/mod/tinygo.org/x/bluetooth@$$bluetooth_ver/s140_nrf52_7.3.0 ./softdevice/

build: softdevice
	@mkdir -p build
	tinygo build $(LD_FLAGS) -tags=$(EXTRA_TAGS) -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=HeadTracker -o ./build/$(FILE) ./src

flash: softdevice
	tinygo flash $(LD_FLAGS) -tags=$(EXTRA_TAGS) -target=$(TARGET) -size=$(SIZE) -opt=z -print-allocs=HeadTracker ./src

monitor:
	tinygo monitor -target=$(TARGET)

# --- Arduino Nano 33 BLE bootloader targets ---

UF2_BOOTLOADER_HEX=./build/arduino_nano_33_ble_bootloader-0.7.0_s140_6.1.1.hex

$(UF2_BOOTLOADER_HEX):
	@curl -L -o $(UF2_BOOTLOADER_HEX) https://github.com/adafruit/Adafruit_nRF52_Bootloader/releases/download/0.7.0/arduino_nano_33_ble_bootloader-0.7.0_s140_6.1.1.hex

flash-uf2-bootloader: $(UF2_BOOTLOADER_HEX)
	nrfjprog -f nrf52 --eraseall
	nrfjprog -f nrf52 --program $(UF2_BOOTLOADER_HEX)

flash-uf2-bootloader-dap: $(UF2_BOOTLOADER_HEX)
	openocd -f interface/cmsis-dap.cfg -f target/nrf52.cfg -c "transport select swd" -c "program $(UF2_BOOTLOADER_HEX) verify reset exit"

# --- Debug targets ---

DEBUG_OPT=1

build-for-debug:
	tinygo build -target=$(TARGET) -tags=$(EXTRA_TAGS) -size=$(SIZE) -opt=$(DEBUG_OPT) -o ./build/debug.elf $(SRC)

debug: build-for-debug
	tinygo gdb -target=$(TARGET) -tags=$(EXTRA_TAGS) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=jlink $(SRC)

debug-dap: build-for-debug
	tinygo gdb -target=$(TARGET) -tags=$(EXTRA_TAGS) -size=$(SIZE) -opt=$(DEBUG_OPT) -ocd-output -programmer=cmsis-dap $(SRC)
