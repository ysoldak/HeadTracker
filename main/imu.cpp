#include "imu.h"

Madgwick filter;

void imuSetup() {

  IMU.begin();
  IMU.setOneShotMode();

  IMU.setAccelFS(2); IMU.setGyroFS(1); IMU.setMagnetFS(0);    // ±4g, ±500°/s, ±400 µT
  IMU.setAccelODR(2); IMU.setGyroODR(2); IMU.setMagnetODR(8); // 50 Hz, 50 Hz, 400Hz

  IMU.setGyroSlope (imuCalibration[0][0], imuCalibration[0][1], imuCalibration[0][2]);
  IMU.setGyroOffset(imuCalibration[1][0], imuCalibration[1][1], imuCalibration[1][2]);

  IMU.setAccelSlope (imuCalibration[2][0], imuCalibration[2][1], imuCalibration[2][2]);
  IMU.setAccelOffset(imuCalibration[3][0], imuCalibration[3][1], imuCalibration[3][2]);

  IMU.setMagnetSlope (imuCalibration[4][0], imuCalibration[4][1], imuCalibration[4][2]);
  IMU.setMagnetOffset(imuCalibration[5][0], imuCalibration[5][1], imuCalibration[5][2]);

  filter.begin(imuSampleFrequency);

}

void imuLagFilter(float_t &curr, float_t &prev) {
  curr = imuLagFilterBeta*curr + (1-imuLagFilterBeta)*prev;
  prev = curr;
}

float_t ax, ay, az, gx, gy, gz, mx, my, mz;
float_t ax1, ay1, az1, gx1, gy1, gz1, mx1, my1, mz1;

void imuUpdate() {

  IMU.readGyro(gx, gy, gz);
  IMU.readAccel(ax, ay, az);
  IMU.readMagnet(mx, my, mz);

  imuLagFilter(gx, gx1); imuLagFilter(gy, gy1); imuLagFilter(gz, gz1);
  imuLagFilter(ax, ax1); imuLagFilter(ay, ay1); imuLagFilter(az, az1);
  imuLagFilter(mx, mx1); imuLagFilter(my, my1); imuLagFilter(mz, mz1);

  filter.update(gx, -gy, gz, ax, -ay, az, -mx, -my, mz); // signs align frames

  // sprintf("%f\t%f\t%f\t\t%f\t%f\t%f\t\t%f\t%f\t%f\n", gx, gy, gz, ax, ay, az, mx, my, mz);
}

float_t imuPan() {
  return filter.getYaw();
}

float_t imuTilt() {
  return filter.getPitch();
}

float_t imuRoll() {
  return filter.getRoll();
}
