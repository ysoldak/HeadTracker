#include "config.h"
#include "imu.h"
#include "para.h"
#include "utils.h"

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
      update();
      millisPrevious += millisPerIter;
    }
  }

  BLE.advertise();
  digitalWrite(BTLED, BTLEDOFF);

}

void imuDebugLoop() {
  millisNow = millis();
  if (millisNow - millisPrevious >= millisPerIter) {
    update();
    if (millisNow%100 == 0) {
      sprintf("%f\t%f\t%f\t|\t%f\t%f\t%f\n", imuPitch(), imuRoll(), imuYaw(), imuStartPitch(), imuStartRoll(), imuStartYaw());
    }
    millisPrevious += millisPerIter;
  }
}

// ----------------------------------------------------------------------------

void update() {

  float_t pitch, roll, yaw;

  if (millisNow - millisConnected < 5000) { // give it 5 sec to settle, record start angle
    imuBegin();
    pitch = 0;
    roll = 0;
    yaw = 0;
  } else {
    imuUpdate();
    pitch = imuPitch();
    roll = imuRoll();
    yaw = imuYaw();
  }

  paraSet(ChannelPitch, toChannel(pitch));
  paraSet(ChannelRoll, toChannel(roll));
  paraSet(ChannelYaw, toChannel(yaw));
  paraSend();

}

uint16_t toChannel(float_t angle) {
  return min(max(round(1500 + 500/AngleMax * angle), 988), 2012);
}
