#include <ArduinoBLE.h>

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
BLECharacteristic fff6("FFF6", BLEWriteWithoutResponse | BLENotify, 32);

uint8_t sysid_data[8] = { 0xF1, 0x63, 0x1B, 0xB0, 0x6F, 0x80, 0x28, 0xFE };
uint8_t m_data[3] = { 0x41, 0x70, 0x70 };
uint8_t ieee_data[14] = { 0xFE, 0x00, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6D, 0x65, 0x6E, 0x74, 0x61, 0x6C };
uint8_t pnpid_data[7] = { 0x01, 0x0D, 0x00, 0x00, 0x00, 0x10, 0x01 };

uint16_t channels[8] = { 1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500 };


void setup() {

  pinMode(LED_BUILTIN, OUTPUT);

  Serial.begin(115200);

  setupBluetooth();

}

void loop() {

  BLEDevice central = BLE.central();

  if (central) {
    Serial.print("Connected to central: ");
    Serial.println(central.address());
    Serial.println(central.localName());

    if (central.connected()) {
      BLE.stopAdvertise();
      digitalWrite(LED_BUILTIN, HIGH);
      if (fff6.subscribe()) {
        Serial.println("Subscribed");
      }
    }

    while (central.connected()) {
      updateChannels();
      sendChannels();
      delay(20);
    }

    Serial.println("Disconnected");
    BLE.advertise();
  }

  digitalWrite(LED_BUILTIN, LOW);
  Serial.print("Waiting for central, our address: ");
  Serial.println(BLE.address());
  delay(100);

}

void setupBluetooth() {
  BLE.begin();
  Serial.println("BLE");

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
  Serial.println(BLE.address());

}

// TODO Read Pan/Tilt from IMU and convert to 1000-2000 range
void updateChannels() {
  channels[3] = 1700;
  channels[6] = 1900;
  channels[7] = 1300;
}

void sendChannels() {
  uint8_t size = encodeChannels();
  // Serial.println(size);
  // for (int i = 0; i<size; i++) {
  //   Serial.print(buffer[i], HEX);
  // }
  // Serial.println();
  fff6.writeValue(buffer, size);
}


// --- Trainer frame encoding ---

uint8_t START_STOP = 0x7E;
uint8_t BYTE_STUFF = 0x7D;
uint8_t STUFF_MASK = 0x20;

uint8_t buffer[32];
uint8_t bufferIndex = 0;
uint8_t crc;

void pushByte(uint8_t byte) {
  crc ^= byte;
  if (byte == START_STOP || byte == BYTE_STUFF) {
    buffer[bufferIndex++] = BYTE_STUFF;
    byte ^= STUFF_MASK;
  }
  buffer[bufferIndex++] = byte;
}

uint8_t encodeChannels() {

  bufferIndex = 0;
  crc = 0x00;

  buffer[bufferIndex++] = START_STOP; // start byte
  pushByte(0x80);             // trainer frame type
  for (int channel=0; channel<8; channel+=2) {
    uint16_t channelValue1 = channels[channel];
    uint16_t channelValue2 = channels[channel+1];
    pushByte(channelValue1 & 0x00ff);
    pushByte(((channelValue1 & 0x0f00) >> 4) + ((channelValue2 & 0x00f0) >> 4));
    pushByte(((channelValue2 & 0x000f) << 4) + ((channelValue2 & 0x0f00) >> 8));
  }
  buffer[bufferIndex++] = crc;
  buffer[bufferIndex++] = START_STOP; // end byte

  return bufferIndex;

}
