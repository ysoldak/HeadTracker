# EthOS 

Full configuration for the HeadTracker and [MicroCameraGimbal](https://cults3d.com/en/3d-model/gadget/micro-camera-gimbal-ysoldak) for [FrSky Tandem X20S](https://www.frsky-rc.com/product/tandem-x20s/) running [EthOS](https://ethos.frsky-rc.com).  

Following assumes you have HT pan servo connected to CH13 and tilt servo to CH14. CH15 is for roll, typically unused.  
HeadTracker runs on XIAO BLE board laying flat on goggles with USB-C facing to the right.  

<img src="../case/DJI%20Integra%20Xiao%20Ble.jpg" width="300">

Note: For HeadTracker mounted some other way `HT TrainerN` mixers will be different, since different trainer channels corespond to different HT axes in that case.  

See also [deep dive article](RadioConfiguration.md) if you want to understand where that % for mixers and outputs come from.

## Mixers

### HT Pot1 (Pan)
```
Name: HT 1
Active condition: Always On
Source: Pot1
Function Type: Add
...
Channels count: 1
Output1: CH13(TR3)
```

### HT Pot2 (Tilt)
```
Name: HT 2
Active condition: Always On
Source: Pot2
Function Type: Add
...
Channels count: 1
Output1: CH14(TR2)
```

### HT Trainer3 (Pan)
```
Name: TR3
Active condition: Always On
Source: Trainer3
Function Type: Add
...
Weight Up: 133%
Weight Down: 133%
...
Channels count: 1
Output1: CH13(TR3)
```
### HT Trainer1 (Tilt)
```
Name: TR1
Active condition: Always On
Source: Trainer1
Function Type: Add
...
Weight Up: 200%
Weight Down: 200%
...
Channels count: 1
Output1: CH14(TR1)
```

### HT Trainer2 (Roll)
```
Name: TR2
Active condition: Always On
Source: Trainer2
Function Type: Add
...
Weight Up: 200%
Weight Down: 200%
...
Channels count: 1
Output1: CH15(TR2)
```


### HT Kill Switch
```
Name: HT
Active condition: SG^
Source: 0
Function Type: Replace
...
Channels count: 3
Output1: CH13(TR3)
Output2: CH14(TR1)
Output3: CH15(TR2)
```

## Outputs

**Attention:** Output ranges must be set in either your Radio or model's Flight Controller, not both!  
If you connect yout pan and tilt servos directly to model's receiver, set output ranges in your radio.  
When pan and tilt servos connected to a Flight Controller, set respective ranges in FC's Outputs tab, leaving radio outputs be 100%.

### CH13 TR3
```
Name: TR3
Min: -150%
Max: 150%
```

## CH14 TR1
```
Name: TR1
Min: -133%
Max: 133%
```
