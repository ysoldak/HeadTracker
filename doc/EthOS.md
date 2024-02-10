# EthOS 

Full configuration for the HeadTracker and [MicroCameraGimbal](https://cults3d.com/en/3d-model/gadget/micro-camera-gimbal-ysoldak) for [FrSky Tandem X20S](https://www.frsky-rc.com/product/tandem-x20s/) running [EthOS](https://ethos.frsky-rc.com).  

Following assumes you have HT pan servo connected to CH13 and tilt servo to CH14. CH15 is for roll, typically unused.  
HeadTracker runs on XIAO BLE board laying flat on goggles with USB-C facing forward. See [Walksnail Goggles X HeadTracker Case](https://www.etsy.com/se-en/listing/1660848137/head-tracker-for-walksnail-avatar).  

Note: For HeadTracker mounted sideways "HT TrainerX" mixers will be different, since different trainer channels corespond to different HT axes in that case.  
See DJI V1/V2 with BDI Adapter and HT mounted in analog bay.

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

### HT Pot2 (Pan)
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

### HT Trainer1 (Roll)
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
Output1: CH15(TR1)
```

### HT Trainer2 (Tilt)
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
Output1: CH14(TR2)
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
Output2: CH14(TR2)
Output3: CH15(TR1)
```

## Outputs

### CH13 TR3
```
Name: TR3
Min: -150%
Max: 150%
```

## CH14 TR2
```
Name: TR2
Min: -130%
Max: 130%
```
