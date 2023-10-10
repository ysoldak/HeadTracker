# How to configure your radio for the head tracker and a gimbal

### TL;DR
_(For the case of this head tracker and [Micro Camera Gimbal](https://cults3d.com/en/3d-model/gadget/micro-camera-gimbal-ysoldak) with Savöx SH-0350 servos)_  

**Output range pan**: 150% (732-2268us)  
**Output range tilt**: 133% (819-2181us)  
**Mixer weight pan**: 133%  
**Mixer weight tilt**: 200%

---

For the best experience with head tracking, you want your HT and camera be always aligned.
This concerns not only drift, but the deflection angle also.
It is quite annoying when you turn your head 90 degrees to the right and camera travels only 60 degrees same direction.
The opposite is annoying too. Let’s try and see how can we configure your radio to keep HT and camera in sync.

## Facts

Typical micro **servo throw** is `90deg` for standard range 988-2012us and `135deg` for extended range 732-2268us.  
Extended range is standard range multiplied by up to 1.5 and can be configured for any channel in most modern radios.  

Another important factor to keep in mind is **gear ratio** (for geared camera gimbals) for pan and tilt.  
For example, [Micro Camera Gimbal](https://cults3d.com/en/3d-model/gadget/micro-camera-gimbal-ysoldak) gear ratios are `2:1` for pan and `3:2` for tilt.


We shall use [Micro Camera Gimbal](https://cults3d.com/en/3d-model/gadget/micro-camera-gimbal-ysoldak)
with [Savöx SH-0350 servos](https://www.savoxusa.com/products/savsh0350-micro-digital-servo-16-36) in our calculations for the rest of this article.

## Calculations

### Head tracker output
The head tracker’s orientation (`-180` to `+180` each axis) is represented by three channels in standard range (988-2012us).  

### Output range
From servo throws for extended output range and gear ratios we conclude gimbal is capable of delivering `270deg` on pan and `202.5deg` on tilt.  
We do not need that much on tilt though and `1.33` output range shall be enough there giving us `180deg`.  

### Mixer weights
Since we have only `270deg` on pan vs. `360deg` that HT delivers, we need to scale trainer (input) channel for pan in radio so gimbal deflection angle matches the one of the head tracker.  
The pan input channel must be scaled by `1.33(=360/270)` and we do that by setting “`Weight:133%`” for respective mixer in radio powered by OpenTX-like firmware.  
Likewise, for tilt the weight value is `200%(=360/180)`.

## Conclusions and notes

That’s all, folks. Hope this helps you configure you radio for the best head tracking experience!

Note though, the exact values may vary a bit depending on which head tracker and camera gimbal you use and what outputs they are capable of.  
Servo throws can vary for brand, model and even individual servos to some extent. The above numbers shall help you started, adjust them for your case!
