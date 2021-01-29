#include "imu.h"
#include "para.h"

uint16_t channels[8] = { 1500, 1500, 1500, 1500, 1500, 1500, 1500, 1500 };

const uint8_t AngleMax    = 45; // for both Pan and Tilt, both directions (x2 total for each axis)
const uint8_t ChannelPan  = 6;
const uint8_t ChannelTilt = 7;

bool debugOutput = true;

unsigned long time_now = 0;

void setup() {

  pinMode(LED_BUILTIN, OUTPUT);
  digitalWrite(LED_BUILTIN, LOW);

  Serial.begin(115200);

  paraSetup();
  imuSetup();

}

void loop() {
  // imuDebugLoop();
  realLoop();
}

void realLoop() {

  while(!BLE.connected()) {
    time_now = millis();
    if (time_now%1000 == 0) {
      Serial.println(BLE.address());
    }
  };

  digitalWrite(LED_BUILTIN, HIGH);
  BLE.stopAdvertise();

  while (BLE.connected()) {
    time_now = millis();
    updateChannels();
    paraSend(channels);
    if (debugOutput && time_now%100 == 0) {
        Serial.print(channels[ChannelPan]); Serial.print(" "); Serial.println(channels[ChannelTilt]);
    }
    delay(5);
  }

  BLE.advertise();
  digitalWrite(LED_BUILTIN, LOW);

}

void imuDebugLoop() {
  updateChannels();
  time_now = millis();
  if (time_now%100 == 0) {
      Serial.print(channels[ChannelPan]); Serial.print(" "); Serial.println(channels[ChannelTilt]);
  }
  delay(5);
}

float_t acc[3], mag[3];
void updateChannels() {
  imuRead(5, acc, mag);
  updatePan(acc, mag);
  updateTilt(acc);
}

// --- Heavy math ---

// Algorithm for Pan in a nutshell:
// - Find true East (E) from Acceleration vector (A), that points down
//   and Magnet vector (M), that points roughly North (N) and down
// - Find true N from E and A
// - Calculate angle between N-E-A vectors and X-Y-Z of the board
//   by projecting N and E on X (or other axis, configurable, depends on board orientation)
// - Apply Kalman(?) filter to reduce jitter

const float_t PanProjectionVector[3] = {1, 0, 0}; // X is along longer edge; Y is along shorter edge
const float_t PanFilterBeta = 0.5; // smaller values dump jitter but also make reaction sluggish

float_t panStartAngle = 0;
float_t panLastAngle = 0;

void updatePan(const float_t acc[3], const float_t mag[3]) {
  float_t E[3];
  float_t N[3];
  vectorCross(acc, mag, E);
  vectorNormalize(E);
  vectorCross(E, acc, N);
  vectorNormalize(N);
  float_t angle = atan2(vectorDot(E, PanProjectionVector), vectorDot(N, PanProjectionVector)) * 180 / PI;
  if (panStartAngle == 0) {
    panStartAngle = angle;
  }
  // return (angle - startPanAngle);
  float_t filteredAngle = (angle - panStartAngle) * PanFilterBeta + (1 - PanFilterBeta) * panLastAngle; // Kalman filter
  panLastAngle = filteredAngle;
  channels[ChannelPan] = toChannel(filteredAngle);
}

void updateTilt(const float_t acc[3]) {
  channels[ChannelTilt] =  toChannel(asin(acc[0]) * 180 / PI);
}

void vectorCross(const float_t a[3], const float_t b[3], float_t out[3]) {
  out[0] = (a[1] * b[2]) - (a[2] * b[1]);
  out[1] = (a[2] * b[0]) - (a[0] * b[2]);
  out[2] = (a[0] * b[1]) - (a[1] * b[0]);
}

float_t vectorDot(const float_t a[3], const float_t b[3]) {
  return (a[0] * b[0]) + (a[1] * b[1]) + (a[2] * b[2]);
}

void vectorNormalize(float_t a[3]) {
  float mag = sqrt(vectorDot(a, a));
  a[0] /= mag; a[1] /= mag; a[2] /= mag;
}

uint16_t toChannel(float_t angle) {
  return min(max(round(1500 + 500/AngleMax * angle), 1000), 2000);
}
