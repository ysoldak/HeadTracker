# Increase throw of an analog servo

Common RC servo has throw of 45 degrees, each side (90 degrees total).
Some servos have total deflection of 120 or even 180 degrees, but they are usually expensive and/or hard to find.

And you generally want at least 180 degrees throw for pan (yaw) camera servo.

With this simple mod you can convert almost any analog 90 degrees servo to 180 degrees servo.
The idea is to tirck the servo contol circuit by increasing resistance of the servo potentiometer. 

Steps:
- Disassemble your servo and locate potentiometer;
- Measure resistance between potentiometer outer pins: `R0`;
- Calculate resistance `R1 = R0 * 0.6`;
- Solder an extra `R1` resistor between servo control circuit and each of outer potentiometer pins;
- Done!


Picture shows two 2.2kOhm resistors soldered to 3.7kOhm potentiometer.
<img src="../media/ServoResistorMod.jpg" title="Servo resistor mod"/>
