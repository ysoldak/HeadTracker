Gyroscope, Accelerometer and Magnetometer sensors must be calibrated prior to use.

For gyroscope it's enough to record offsets while board is stationary and compensate for that later.

For accelerometer you want to allign all three axes perpendicular to the ground (both up and down), 
note readings and fairly simply calculate offsets and scales/gains and compensate for that.

Magnetometer calibration is harder.

Here is some material to understand the process.  
Information overlaps, still keeping all relevant links for completenes.

- [Compensating for Tilt, Hard-Iron, and Soft-Iron Effects](https://www.fierceelectronics.com/components/compensating-for-tilt-hard-iron-and-soft-iron-effects)
- [FAQ: Hard & Soft Iron Correction for Magnetometer Measurements](https://ez.analog.com/mems/w/documents/4493/faq-hard-soft-iron-correction-for-magnetometer-measurements)
- [Magnetometer Errors & Calibration](https://www.vectornav.com/resources/magnetometer-errors-calibration)

Other useful links:
- [Circle and Elipse visualisation](https://www.desmos.com/calculator/p52mkrcvrm)
- [Axes conventions](https://en.wikipedia.org/wiki/Axes_conventions)
