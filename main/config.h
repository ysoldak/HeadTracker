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

  // board 1 (grey)
  // { 1.00,  1.00,  1.00},
  // { 1.15,  -0.05,  -0.15},
  // { 0.99,  1.00,  0.99},
  // { -0.07, -0.01, 0.00},
  // { 1.20,  1.18,  1.20},
  // {15.00, 19.00, 0.00}

  // board 2 (yellow)
  // { 1.00,  1.00,  1.00 },
  // { 0.70,  0.25,  1.55 },
  // { 1.00,  1.00,  0.99 },
  // { -0.01, -0.02,  0.05 },
  // { 1.17,  1.17,  1.23 },
  // { -8.00, 37.00,  0.50 }
};

// Max angle for Pan, Tilt and Roll; each direction (x2 total for each axis)
const uint8_t AngleMax = 45;

// Trainer channels Head Tracker sends data
const uint8_t ChannelPan  = 0;
const uint8_t ChannelTilt = 1;
const uint8_t ChannelRoll = 2;

const uint8_t BoardTilt = 0;
const uint8_t BoardRoll = 90;

// -- Knobs to turn if you really know what you are doing ---

const float_t imuSampleFrequency = 50;
const float_t imuLagFilterBeta = 0.5;

const unsigned long millisPerIter = 1000/imuSampleFrequency;

const pin_size_t BTLED = LEDB;
const PinStatus BTLEDON = LOW;
const PinStatus BTLEDOFF = HIGH; // for RGB LED "high" means OFF

const bool htDebug = false;
