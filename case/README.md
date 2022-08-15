## Xiao BLE Analog Bay Case 1

Print `XiaoBle1.stl` standing on USB-C connector hole side.

The case is designed for Xiao BLE board and SSD1306 128x32 display.  
Use double-size tape or a drop of rubber glue to attach display to Xiao BLE.
Insert USB-C of Xiao BLE to respective hole and press the display in place. Use a drop of hot glue or a piece tape to secure the display on place.

Optionally, drill a hole and install the reset orientation button connected to `D2` and `GND`.

<table>
<tr><td>
<img src="XiaoBle1Closed.jpg" title="XIAO + SSD1306 128x32 on DJI Goggles" style="float: left;"/>
</td><td>
<img src="XiaoBle1Open.jpg" title="XIAO + SSD1306 128x32 on DJI Goggles, wiring" style="float: right;"/>
</td></tr>
</table>


## Xiao BLE Analog Bay Case 2

There is also an option to place Xiao BLE board with display on the right side of Fatshark googles.  
Power shall be sourced from googles themselves, 3.3v is available there.

Please check `XiaoBle2.3mf` file and pictures for placement and wiring hints.

<table>
<tr><td>
<img src="XiaoBle2Closed.jpg" title="XIAO + SSD1306 128x32 on Fatshark Goggles" style="float: left;"/>
</td><td>
<img src="XiaoBle2Open.jpg" title="XIAO + SSD1306 128x32 on Fatshark Goggles, wiring" style="float: right;"/>
</td></tr>
</table>


## Nano 33 BLE Side Case

Print `Nano33Ble-Main.stl` on its back, `Nano33Ble-Clip.stl` standing.

`Nano33Ble-Main.stl` has a hole above RGB LED (indicates Bluetooth connection state).  
Insert a piece of transparent fillament into it, for a light guide.

Glue a piece of little something on the case's reboot button (from inside or outside, your preference) to locate it easier.  
Board records orientation on boot, so rebooting is a way to reset orientation, no additional buttons needed.  
Note: Bluetooth connection will break on reboot just to restore quickly after.

Solder 2S balance plug to the board and press-fit it into case -- can power the board from FatShark's battery.  
There is a place for on/off switch, but that's optional.

<table>
<tr><td>
<img src="Nano33BleOnGoggles.jpg" title="Nano 33 BLE on FatShark mounted on the left side" style="float: left;"/>
</td><td>
<img src="Nano33BleOpen.jpg" title="Nano 33 BLE on FatShark, wiring" style="float: right;"/>
</td></tr>
</table>
