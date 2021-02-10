#include "config.h"
#include "imu.h"
#include "para.h"
#include "utils.h"

uint16_t channels[8] = { 1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500 };

unsigned long millisNow = 0;
unsigned long millisPrevious = 0;
unsigned long millisConnected = 0;

float_t startAngle = 0.0;

// ----------------------------------------------------------------------------

void setup() {

  pinMode(BTLED, OUTPUT);
  digitalWrite(BTLED, BTLEDOFF);

  Serial.begin(115200);

  paraSetup();
  imuSetup();

}

void loop() {
  // imuDebugLoop();
  realLoop();
}

// ----------------------------------------------------------------------------

void realLoop() {

  while(!BLE.connected()) {
    millisNow = millis();
    if (millisNow%1000 == 0) {
      Serial.println(BLE.address());
      digitalWrite(BTLED, BTLEDON);
    } else if (millisNow%500 == 0) {
      digitalWrite(BTLED, BTLEDOFF);
    }
  };

  digitalWrite(BTLED, BTLEDON);
  BLE.stopAdvertise();
  millisConnected = millis();

  while (BLE.connected()) {
    millisNow = millis();
    if (millisNow - millisPrevious >= millisPerIter) {
      updateChannels();
      paraSend(channels);
      millisPrevious = millisPrevious + millisPerIter;
    }
  }

  BLE.advertise();
  digitalWrite(BTLED, BTLEDOFF);

}

void imuDebugLoop() {
  millisNow = millis();
  if (millisNow - millisPrevious >= millisPerIter) {
    updateChannels();
    millisPrevious = millisPrevious + millisPerIter;
  }
}

// ----------------------------------------------------------------------------

void updateChannels() {

  imuUpdate();

  float_t pan  = imuPan();
  float_t tilt = imuTilt();
  float_t roll = imuRoll();

  if (millisNow - millisConnected < 5000) { // give it 5 sec to settle, record start angle
    startAngle = pan;
    tilt = 0;
    roll = 0;
  }
  if (startAngle < 180 && pan > startAngle + 180) {
    pan -= 360;
  } else if (startAngle > 180 && pan < startAngle - 180) {
    pan += 360;
  }
  pan -= startAngle;

  channels[ChannelPan]  = toChannel(pan);
  channels[ChannelTilt] = toChannel(tilt);
  channels[ChannelRoll] = toChannel(roll);

  if (htDebug && millisNow%100 == 0) {
    sprintf("%f\t%f\t%f\t%f\n", startAngle, pan, tilt, roll);
  }

}

uint16_t toChannel(float_t angle) {
  return min(max(round(1500 + 500/AngleMax * angle), 1000), 2000);
}
