#pragma once

// Calibration matrix for Gyroscope, Accelerometer and Magnetometer
// Columns: x, y, z
const float_t imuCalibration[6][3] = {
  {1, 1, 1}, // gyroscope gain
  {0, 0, 0}, // gyroscope offset
  {1, 1, 1}, // accelerometer gain
  {0, 0, 0}, // accelerometer offset
  {1, 1, 1}, // magnetometer gain    -- soft iron
  {0, 0, 0}  // magnetometer offset  -- hard iron

  // board 1
  // { 1.00,  1.00,  1.00},
  // { 1.12,  0.07,  0.00},
  // { 0.99,  0.99,  0.99},
  // { 0.00,  0.00,  0.00},
  // { 1.16,  1.12,  1.14},
  // {12.11, 24.21, -2.76}

  // board 2
  // { 1.00,  1.00,  1.00 },
  // { 0.70,  0.10, -2.30 },
  // { 1.00,  1.00,  1.00 },
  // { 0.00,  0.00,  0.00 },
  // { 1.11,  1.12,  1.17 },
  // { 3.50, 24.60,  0.30 }
};

// Max angle for Pan, Tilt and Roll; each direction (x2 total for each axis)
const uint8_t AngleMax = 45;

// Trainer channels Head Tracker sends data
const uint8_t ChannelPan  = 0;
const uint8_t ChannelTilt = 1;
const uint8_t ChannelRoll = 2;


// -- Knobs to turn if you really know what you are doing ---

const float_t imuSampleFrequency = 50;
const float_t imuLagFilterBeta = 0.5;

const unsigned long millisPerIter = 1000/imuSampleFrequency;

const pin_size_t BTLED = LEDB;
const PinStatus BTLEDON = LOW;
const PinStatus BTLEDOFF = HIGH; // for RGB LED "high" means OFF

const bool htDebug = false;
