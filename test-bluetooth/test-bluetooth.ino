#include <ArduinoBLE.h>

// Minimal sketch to test Bluetooth communication with OpenTX radio.
// Custom ArduinoBLE library required, download it from https://github.com/ysoldak/ArduinoBLE/tree/cccd_hack
// Diff: https://github.com/ysoldak/ArduinoBLE/compare/master...ysoldak:cccd_hack

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
BLECharacteristic fff6("FFF6", BLEWriteWithoutResponse | BLENotify | BLEAutoSubscribe, 32);

uint8_t sysid_data[8] = { 0xF1, 0x63, 0x1B, 0xB0, 0x6F, 0x80, 0x28, 0xFE };
uint8_t m_data[3] = { 0x41, 0x70, 0x70 };
uint8_t ieee_data[14] = { 0xFE, 0x00, 0x65, 0x78, 0x70, 0x65, 0x72, 0x69, 0x6D, 0x65, 0x6E, 0x74, 0x61, 0x6C };
uint8_t pnpid_data[7] = { 0x01, 0x0D, 0x00, 0x00, 0x00, 0x10, 0x01 };


void setup() {

  pinMode(LED_BUILTIN, OUTPUT);

  Serial.begin(115200);

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

void loop() {

  BLEDevice central = BLE.central();

  if (central) {
    Serial.print("Connected to central: ");
    Serial.println(central.address());
    Serial.println(central.localName());

    if (central.connected()) {
      BLE.stopAdvertise();
      digitalWrite(LED_BUILTIN, HIGH);
    }

    while (central.connected()) {
      sendFakeTrainer();
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

// THR up: 7e 80 dd 5d c5 dc 7d 5d c5 dc 5d c5 dc 5d c5 a1 7e
void sendFakeTrainer() {
  uint8_t data[17] = { 0x7E, 0x80, 0xDD, 0x5D, 0xC5, 0xDC, 0x7D, 0x5D, 0xC5, 0xDC, 0x5D, 0xC5, 0xDC, 0x5D, 0xC5, 0xA1, 0x7E };
  fff6.writeValue(data, 17);
}
