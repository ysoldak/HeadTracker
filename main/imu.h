#pragma once

#include <Arduino_LSM9DS1.h>
#include <MadgwickAHRS.h>

#include "config.h"
#include "utils.h"

void imuSetup();
void imuBegin();
void imuUpdate();

float_t imuPitch();
float_t imuRoll();
float_t imuYaw();

float_t imuStartPitch();
float_t imuStartRoll();
float_t imuStartYaw();
