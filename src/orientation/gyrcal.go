package orientation

// Constant gyroscope calibration for slow moving objects.
//
// Every gyroscope has bias and jitter on each of its 3 axes.
// The bias is usually relatively small but induces drift with time, if not adjusted for.
//
// The algorithm implemented here is suited for slowly rotating objects like a head tracker.
// Calibration addresses bias only, represented with "Offset" for each of axes.
//
// When gyroscope is calibrated and stationary, values from gyro with offsets applied, shall all be around zero.
// The non-zero small values are jitter that shall compensate itself out when summed up.
//
// When gyroscope is stationary, or nearly stationary, we can observe bias on each of axis by
// summing up read values and taking avg value of them ("gyrCalBatchSize").
// The offset then adjusted to minimise next calculated avg value.
//
// Note:
// We can not expect the device be stationary for a long time, we also want to keep calibarion running
// even during normal operation. The non-stationary values are recognised when they escape "gyrCalEscapeThreshold".
// A whole batch of values is ignored when many large values are observed, see "gyrCalBatchEscapeMaxCount".
//
// The calibration is good enough when latest offset correction for each of axes are small,
// that means remaining error is small too and can not induce much drift anymore.
// The good enough calibration is indicated by "Stable" flag.

const (
	gyrCalBatchSize            = 1000                      // 1 sec on warm-up, 20 sec during regular operation
	gyrCalBatchEscapeMaxCount  = gyrCalBatchSize / 100 * 3 // tolerate 3% values outside threshold
	gyrCalForceBatchesCount    = 10                        // first 10 batches applied always, regardles of number of escapes
	gyrCalEscapeThreshold      = 4_000_000                 // this is hardware center point precision (we can expect values in this range when stationary)
	gyrCalStableCheckThreshold = 100_000                   // shall be small enough to eliminate axis drift while large enough to keep time of warm-up low
	gyrCalStableCorrThreshold  = 25_000                    // when stable, corrections can be large only when induced by fast movements, so capping them
)

type GyrCal struct {
	Stable bool     // reached good enough offsets at least once
	Offset [3]int32 // current calibration offsets

	correctionLast [3]int32
	correctionSum  [3]int32

	countApply  [3]int32
	countEscape [3]int32
	countForce  [3]int32
}

func (g *GyrCal) Get(x, y, z int32) (int32, int32, int32) {
	return x - g.Offset[0], y - g.Offset[1], z - g.Offset[2]
}

func (g *GyrCal) Apply(x, y, z int32) {
	g.applyAxis(0, x)
	g.applyAxis(1, y)
	g.applyAxis(2, z)
	if g.Stable {
		return
	}
	g.Stable = true
	for i := 0; i < 3; i++ {
		g.Stable = g.Stable && g.correctionLast[i] != 0 && abs(g.correctionLast[i]) < gyrCalStableCheckThreshold
	}
}

func (g *GyrCal) applyAxis(i, v int32) {
	value := v - g.Offset[i]
	if abs(value) > gyrCalEscapeThreshold {
		value = sign(value) * gyrCalEscapeThreshold
		g.countEscape[i]++
	}
	g.correctionSum[i] += value / gyrCalBatchSize // divide right away to avoid integer overflow
	g.countApply[i]++

	if g.countApply[i] > gyrCalBatchSize {
		g.adjustAxisOffset(i)
	}

}

func (g *GyrCal) adjustAxisOffset(i int32) {
	// adjust when relatively stable or first times
	if g.countEscape[i] < gyrCalBatchEscapeMaxCount || g.countForce[i] < gyrCalForceBatchesCount {
		g.correctionLast[i] = g.correctionSum[i] / 2                          // be careful and half-step
		if g.Stable && abs(g.correctionLast[i]) > gyrCalStableCorrThreshold { // cap correction when stable
			g.correctionLast[i] = sign(g.correctionLast[i]) * gyrCalStableCorrThreshold
		}
		g.Offset[i] += g.correctionLast[i]
		if g.countForce[i] < gyrCalForceBatchesCount {
			g.countForce[i]++
		}
	}
	g.correctionSum[i] = 0
	g.countApply[i] = 0
	g.countEscape[i] = 0
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int32) int32 {
	if v == 0 {
		return 0
	}
	if v > 0 {
		return 1
	}
	return -1
}
