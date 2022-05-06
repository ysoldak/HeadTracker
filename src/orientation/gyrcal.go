package orientation

const (
	gyrCalBatchSize           = 1000
	gyrCalValueThreshold      = 3_000_000 // this is hardware center point precision
	gyrCalCorrectionThreshold = 100_000
)

type GyrCal struct {
	Stable bool     // indicates reached good enough offsets at least once
	Offset [3]int32 // current calibration offsets

	lastCorrection [3]int32
	valuesSum      [3]int32
	valuesCount    [3]int32
	escapeCount    [3]int32
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
		if g.lastCorrection[i] == 0 || abs(g.lastCorrection[i]) > gyrCalCorrectionThreshold {
			g.Stable = false
			break
		}
	}
}

func (g *GyrCal) applyAxis(i, v int32) {
	tmp := v - g.Offset[i]
	if abs(tmp) > gyrCalValueThreshold {
		g.escapeCount[i]++
		tmp = sign(tmp) * gyrCalValueThreshold
	}
	g.valuesSum[i] += tmp
	g.valuesCount[i]++
	if g.valuesCount[i] > gyrCalBatchSize {
		if g.escapeCount[i] < gyrCalBatchSize/10 || g.Offset[i] == 0 { // only adjust if was relatively stable or the very first time
			g.lastCorrection[i] = g.valuesSum[i] / g.valuesCount[i] / 2
			g.Offset[i] += g.lastCorrection[i]
		} else { // else discard
			g.lastCorrection[i] = 0
		}
		g.valuesCount[i] = 0
		g.valuesSum[i] = 0
		g.escapeCount[i] = 0
	}
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
