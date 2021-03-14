#include "para.h"

// --- Bluetooth PARA trainer connection ---

// Here we try and mimic FrSky radio as much as possible
// Probably something can be dropped
BLEService info("180A");
BLECharacteristic sysid("2A23", BLERead, 8);
BLECharacteristic manufacturer("2A29", BLERead, 3);
BLECharacteristic ieee("2A2A", BLERead, 14);
BLECharacteristic pnpid("2A50", BLERead, 7);

BLEService para("FFF0");
BLEByteCharacteristic fff1("FFF1", BLERead | BLEWrite);
BLEByteCharacteristic fff2("FFF2", BLERead);
BLECharacteristic fff3("FFF3", BLEWriteWithoutResponse, 32);
BLECharacteristic fff5("FFF5", BLERead, 32);

// FFF6 characteristic is communication channel with radio.
// Seems like FrSky radio either does not subscribe to notifications properly or we fail to notice it.
// So we force-enable notifications (BLEAutoSubscribe) -- patched version of ArduinoBLE is needed for that.
// Patch ensures CCCD for FFF6 characteristic stays enabled (0x0001)
BLECharacteristic fff6("FFF6", BLEWriteWithoutResponse | BLENotify | BLEAutoSubscribe, 32);

uint8_t sysid_data[8] = { 0xF1, 0x63, 0x1B, 0xB0, 0x6F, 0x80, 0x28, 0xFE };
uint8_t m_data[3] = { 0x41, 0x70, 0x70 };
uint8_t ieee_data[14] = { 0xFE, 0x00, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6D, 0x65, 0x6E, 0x74, 0x61, 0x6C };
uint8_t pnpid_data[7] = { 0x01, 0x0D, 0x00, 0x00, 0x00, 0x10, 0x01 };

uint16_t channels[8] = { 1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500 };

void paraSetup() {

  Serial.print("Bluetooth");

  BLE.begin();

  BLE.setConnectable(true);
  BLE.setLocalName("Hello");

  info.addCharacteristic(sysid);
  info.addCharacteristic(manufacturer);
  info.addCharacteristic(ieee);
  info.addCharacteristic(pnpid);
  BLE.addService(info);
  sysid.writeValue(sysid_data, 8);
  manufacturer.writeValue(m_data, 3);
  ieee.writeValue(ieee_data, 14);
  pnpid.writeValue(pnpid_data, 7);

  BLE.setAdvertisedService(para);
  para.addCharacteristic(fff1);
  para.addCharacteristic(fff2);
  para.addCharacteristic(fff3);
  para.addCharacteristic(fff5);
  para.addCharacteristic(fff6);

  BLE.addService(para);
  fff1.writeValue(0x01);
  fff2.writeValue(0x02);

  BLE.advertise();

  Serial.println(" +");
  Serial.println(BLE.address());

}

void paraSet(uint8_t channel, uint16_t value) {
  channels[channel] = value;
}

const uint8_t START_STOP = 0x7E;
const uint8_t BYTE_STUFF = 0x7D;
const uint8_t STUFF_MASK = 0x20;

// Escapes bytes that equal to START_STOP and updates CRC
void paraPushByte(uint8_t byte, uint8_t* buffer, uint8_t& bufferIndex, uint8_t& crc) {
  crc ^= byte;
  if (byte == START_STOP || byte == BYTE_STUFF) {
    buffer[bufferIndex++] = BYTE_STUFF;
    byte ^= STUFF_MASK;
  }
  buffer[bufferIndex++] = byte;
}

// Encodes channels array to para trainer packet (adapted from OpenTX source code)
void paraSend() {

  uint8_t buffer[32];
  uint8_t bufferIndex = 0;
  uint8_t crc = 0x00;

  buffer[bufferIndex++] = START_STOP;            // start byte
  paraPushByte(0x80, buffer, bufferIndex, crc);  // trainer frame type
  for (int channel=0; channel<8; channel+=2) {
    uint16_t channelValue1 = channels[channel];
    uint16_t channelValue2 = channels[channel+1];
    paraPushByte(channelValue1 & 0x00ff, buffer, bufferIndex, crc);
    paraPushByte(((channelValue1 & 0x0f00) >> 4) + ((channelValue2 & 0x00f0) >> 4), buffer, bufferIndex, crc);
    paraPushByte(((channelValue2 & 0x000f) << 4) + ((channelValue2 & 0x0f00) >> 8), buffer, bufferIndex, crc);
  }
  buffer[bufferIndex++] = crc;                   // crc byte
  buffer[bufferIndex++] = START_STOP;            // end byte

  fff6.writeValue(buffer, bufferIndex);
}
