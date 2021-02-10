#pragma once

#include <Arduino_LSM9DS1.h>
#include <MadgwickAHRS.h>

#include "config.h"
#include "utils.h"

void imuSetup();
void imuUpdate();

float_t imuPan();
float_t imuTilt();
float_t imuRoll();
