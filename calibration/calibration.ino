#include <Arduino_LSM9DS1.h>
#include <stdarg.h>

// Please, set these two constants for your location
// https://en.wikipedia.org/wiki/Earth%27s_magnetic_field

const float_t EarthMagnetStrength    = 51.5; //= µT (Stockholm)
const float_t EarthMagnetInclination = 73;   //=deg (Stockholm)

// Following constants shall be good enough, tweak if feeling adventurous

const uint8_t NofCalibrationSamples = 10;  // Number of samples collected for each axis considered enough to start calibration
const float_t AccelCriterion        = 0.1; // Accelerometer axis detection tolerance


// Expected value of magnetic field strength along each axis
const float_t MagTarget = EarthMagnetStrength * cos((90-EarthMagnetInclination) * PI/180);

const char xyz[3]= {'X','Y','Z'};

float_t imuCalibration[4][3] = {
  {1, 1, 1},
  {0, 0, 0},
  {1, 1, 1},
  {0, 0, 0}
};


float_t calSamplesMin[3] = {0, 0, 0};
float_t calSamplesMax[3] = {0, 0, 0};
float_t magSamplesMin[3] = {0, 0, 0};
float_t magSamplesMax[3] = {0, 0, 0};


void setup() {
  Serial.begin(115200);
  while (!Serial);
  delay(10);
  IMU.begin();
  IMU.setOneShotMode();
  IMU.setAccelFS(2); IMU.setGyroFS(1); IMU.setMagnetFS(0);    // ±4g, ±500°/s, ±400 µT
  IMU.setAccelODR(2); IMU.setGyroODR(2); IMU.setMagnetODR(8); // 50 Hz, 50 Hz, 400Hz
}

void loop() {
  calibrateMenu();
}

void calibrateMenu() {
  while (1) {
    Serial.println(F("\n\n\n\n\n\n\n\n\n\n\n"));
    Serial.println(F("Calibrate Accelerometer and Magnetometer\n"));

    Serial.println(F(" Run calibration a couple of times for the values to settle\n"));
    Serial.println(F(" and errors [shown in square brackets] be in single digits\n"));

    Serial.println(F(" (Enter)  Start calibration"));

    Serial.println("\n\n");
    for (int i = 0; i < 4; i++) {
      Serial.print("{");
      for (int j = 0; j < 3; j++) {
        Serial.print(imuCalibration[i][j]); if (j < 2) Serial.print(", ");
      }
      Serial.print("}");
      if (i < 3) Serial.print(",");
      Serial.println();
    }
    Serial.println(F("\n\n"));

    readChar();
    calibrate();
  }
}


char readChar()
{  char ch;
   while (!Serial.available()) ;             // wait for character to be entered
   ch= toupper(Serial.read());
   delay(10);
   while (Serial.available()){Serial.read();delay(1);} // empty readbuffer
   return ch;
}

// TODO: Calibrate accelerometer
void calibrate() {
  Serial.println(F("\n\n\n\n\n\n\n\n\n\n\n"));

  Serial.println(F(" Place the board on a horizontal surface with one of its axes vertical and wait for confirmation."));
  Serial.println(F(" Each of the axes must be measured pointing up and pointing down, so a total of 6 measurements."));
  Serial.println(F(" The program recognises which axis is vertical and shows which were measured successfully."));
  Serial.println(F(" If the angle is to far oblique the measurement is not valid.\n  "));
  Serial.println(F(" The magnetic field measurement will be heavily disturbed by your set-up, so an \"in-situ\" calibration is advised.\n"));

  Serial.println(F("\n\nCalibrating..."));

  bool enough = false;

  while(!enough) {
    float_t acc[3];
    float_t mag[3];
    readAvg(10, acc, mag);

    bool shallPrintStatus = false;
    for (int i = 0; i<3; i++) {
      float_t a = acc[i], b = acc[(i+1)%3], c = acc[(i+2)%3];
      if (abs(a)>max(abs(b),abs(c))) {
        if (sqrt(b*b+c*c)/abs(a)<AccelCriterion) {
          Serial.print(". ");
          if (a < 0) {
            magSamplesMin[i] = min(mag[i], magSamplesMin[i]);
            calSamplesMin[i]++;
          } else {
            magSamplesMax[i] = max(mag[i], magSamplesMax[i]);
            calSamplesMax[i]++;
          }
          shallPrintStatus = (calSamplesMin[i] == NofCalibrationSamples) || (calSamplesMax[i] == NofCalibrationSamples);
          break;
        }
      }
    }
    if (shallPrintStatus) {
      Serial.println();
      for (int i=0; i<3; i++) {
        Serial.print(xyz[i]);Serial.print("-");
        if (calSamplesMin[i] >= NofCalibrationSamples) {
          Serial.print("OK [ "); Serial.print(MagTarget+magSamplesMin[i]); Serial.print(" ]   ");
        } else {
          Serial.print(" -- [ --.-- ]   ");
        }
        Serial.print(xyz[i]);Serial.print("+");
        if (calSamplesMax[i] >= NofCalibrationSamples) {
          Serial.print("OK [ "); Serial.print(magSamplesMax[i]-MagTarget); Serial.print(" ]   ");
        } else {
          Serial.print(" -- [ --.-- ]   ");
        }
      }
      Serial.println();
    }

    //  float_t cosValue = vector_dot(mag, acc) / (sqrt(vector_dot(mag, mag)) * sqrt(vector_dot(acc, acc)));
    //  float_t angle = 90 - acos (cosValue) * 180.0 / PI;
    //  printFormatted("Acc: %f,%f,%f    Mag: %f,%f,%f    Angle: %f", acc[0], acc[1], acc[2], mag[0], mag[1], mag[2], angle);
    //  printFormatted("%f, %f, %f / %f, %f, %f  |  %f, %f, %f / %f, %f, %f\n", magSamplesMin[0], magSamplesMin[1], magSamplesMin[2], magSamplesMax[0], magSamplesMax[1], magSamplesMax[2], calSamplesMin[0], calSamplesMin[1], calSamplesMin[2], calSamplesMax[0], calSamplesMax[1], calSamplesMax[2]);

    enough = true;
    for (int i = 0; i<3; i++) {
      if (magSamplesMin[i] == 0 || magSamplesMax[i] == 0 || calSamplesMin[i] < NofCalibrationSamples || calSamplesMax[i] < NofCalibrationSamples) {
        enough = false;
        break;
      }
    }

  }

  float_t offsetX = imuCalibration[3][0]+(magSamplesMax[0]+magSamplesMin[0])/2/imuCalibration[2][0];
  float_t offsetY = imuCalibration[3][1]+(magSamplesMax[1]+magSamplesMin[1])/2/imuCalibration[2][1];
  float_t offsetZ = imuCalibration[3][2]+(magSamplesMax[2]+magSamplesMin[2])/2/imuCalibration[2][2];
  float_t slopeX  = imuCalibration[2][0]*(2*MagTarget)/(magSamplesMax[0]-magSamplesMin[0]);
  float_t slopeY  = imuCalibration[2][1]*(2*MagTarget)/(magSamplesMax[1]-magSamplesMin[1]);
  float_t slopeZ  = imuCalibration[2][2]*(2*MagTarget)/(magSamplesMax[2]-magSamplesMin[2]);
  imuCalibration[2][0] = slopeX;
  imuCalibration[2][1] = slopeY;
  imuCalibration[2][2] = slopeZ;
  imuCalibration[3][0] = offsetX;
  imuCalibration[3][1] = offsetY;
  imuCalibration[3][2] = offsetZ;

  for (int i = 0; i<3; i++) {
    magSamplesMin[i] = 0;
    magSamplesMax[i] = 0;
    calSamplesMin[i] = 0;
    calSamplesMax[i] = 0;
  }

}


void readAvg(uint8_t N, float_t a[3], float_t m[3]) {

  for (uint8_t i = 0; i < 3; i++) {
    a[i] = 0;
    m[i] = 0;
  }

  float_t x, y, z;
  for (uint8_t i = 0; i < N; i++) {
    while (!IMU.magneticFieldAvailable() || !IMU.accelerationAvailable());
    IMU.readRawAccel(x, y, z);
    a[0] += x;
    a[1] -= y;
    a[2] -= z; // substract on Y and Z to align acc vector with mag vector
    IMU.readRawMagnet(x, y, z);
    m[0] += imuCalibration[2][0] * (x - imuCalibration[3][0]);
    m[1] += imuCalibration[2][1] * (y - imuCalibration[3][1]);
    m[2] += imuCalibration[2][2] * (z - imuCalibration[3][2]);
  }

  for (uint8_t i = 0; i < 3; i++) {
    a[i] /= N;
    m[i] /= N;
  }

}


// - output utils -------------------------------------------------------------------

void printParam(char txt[], float param[3]){
  for (int i= 0; i<=2 ; i++) {
    Serial.print(txt);Serial.print("[");
    Serial.print(i);Serial.print("] = ");
    Serial.print(param[i],6);Serial.print(";");
  }
}

void printSetParam(char txt[], float param[3]) {
  Serial.print(txt);Serial.print("(");
  Serial.print(param[0],6);Serial.print(", ");
  Serial.print(param[1],6);Serial.print(", ");
  Serial.print(param[2],6);Serial.print(");");
}

void printFormatted(const char* input...) {
  va_list args;
  va_start(args, input);
  for(const char* i=input; *i!=0; ++i) {
    if(*i!='%') { Serial.print(*i); continue; }
    switch(*(++i)) {
      case '%': Serial.print('%'); break;
      case 's': Serial.print(va_arg(args, char*)); break;
      case 'd': Serial.print(va_arg(args, int), DEC); break;
      case 'b': Serial.print(va_arg(args, int), BIN); break;
      case 'o': Serial.print(va_arg(args, int), OCT); break;
      case 'x': Serial.print(va_arg(args, int), HEX); break;
      case 'f': Serial.print(va_arg(args, double), 2); break;
    }
  }
  va_end(args);
}

// - math utils -------------------------------------------------------------------

void vector_cross(const float_t a[3], const float_t b[3], float_t out[3]) {
  out[0] = (a[1] * b[2]) - (a[2] * b[1]);
  out[1] = (a[2] * b[0]) - (a[0] * b[2]);
  out[2] = (a[0] * b[1]) - (a[1] * b[0]);
}

float vector_dot(const float_t a[3], const float_t b[3]) {
  return (a[0] * b[0]) + (a[1] * b[1]) + (a[2] * b[2]);
}

void vector_normalize(float_t a[3]) {
  float mag = sqrt(vector_dot(a, a));
  a[0] /= mag; a[1] /= mag; a[2] /= mag;
}
