#include "imu.h"

// acc scale, acc offset, mag scale, mag offset; one value per axis: x, y, z
float_t imuCalibration[4][3] = {
  {1, 1, 1},
  {0, 0, 0},
  {1, 1, 1},
  {0, 0, 0}
};

// specific board calibration results
float_t imuCalibrationBoard_7BF5[4][3] = {
  { 0.99,  0.99,  1.00},
  {-0.01, -0.01, -0.01},
  { 1.35,  1.32,  1.38},
  {17.57, 21.24, -0.21}
};



// ---

void imuSetup() {
  IMU.begin();
  // Disables accelerometer FIFO (magnetometer has no buffer).
  // This setup consumes more power (not an issue in our case)
  // but provides more accurate results
  IMU.setOneShotMode();
  // Load calibration data for _this_ board
  memcpy(imuCalibration, imuCalibrationBoard_7BF5, sizeof(imuCalibration));
}

// Reads both vectors once
void imuRead(float_t acc[3], float_t mag[3]) {
  float_t values[3];
  imuReadAcceleration(values);
  for (int j = 0; j < 3; j++) {
    acc[j] = imuCalibration[0][j] * ( values[j] - imuCalibration[1][j] );
  }
  imuReadMagneticField(values);
  for (int j = 0; j < 3; j++) {
    mag[j] = imuCalibration[2][j] * ( values[j] - imuCalibration[3][j] );
  }
}

// Reads both vectors N times and averages
void imuRead(int N, float_t acc[3], float_t mag[3]) {

  for (int i = 0; i < 3; i++) {
    acc[i] = 0;
    mag[i] = 0;
  }

  float_t acc1[3], mag1[3];
  for (int n = 0; n < N; n++) {
    imuRead(acc1, mag1);
    for (int i = 0; i < 3; i++) {
      acc[i] += acc1[i];
      mag[i] += mag1[i];
    }
  }

  for (int i = 0; i < 3; i++) {
    acc[i] /= N;
    mag[i] /= N;
  }

}

// Inverts on Y and Z to align acc vector with mag vector
void imuReadAcceleration(float_t out[3]) {
  float_t x, y, z;
  while (!IMU.accelerationAvailable());
  IMU.readAcceleration(x, y, z);
  out[0] = x; out[1] = -y; out[2] = -z;
}

// Filters magnetic field value
void imuReadMagneticField(float_t out[3]) {
  float_t variance[3];
  imuReadMagneticField(out, variance);
}

float_t filterMagBeta = 0.25;
float_t filterMagPrev[3] = {0, 0, 0};

// Just in case one wants to see variance (jitter)
void imuReadMagneticField(float_t out[3], float_t variance[3]) {
  while (!IMU.magneticFieldAvailable());
  IMU.readMagneticField(out[0], out[1], out[2]);
  for (int i = 0; i < 3; i++) {
    if (filterMagPrev[i] == 0) {
      filterMagPrev[i] = out[i];
    }
    out[i] = out[i] * filterMagBeta + (1 - filterMagBeta) * filterMagPrev[i];
    variance[i] = out[i] - filterMagPrev[i];
    filterMagPrev[i] = out[i];
  }
}
