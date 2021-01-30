#pragma once

#include <Arduino_LSM9DS1.h>

void imuSetup();

void imuRead(float_t acc[3], float_t mag[3]);
void imuRead(int N, float_t acc[3], float_t mag[3]);

void imuReadAcceleration(float_t out[3]);
void imuReadMagneticField(float_t out[3]);
void imuReadMagneticField(float_t out[3], float_t variance[3]);
