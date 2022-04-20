package orientation

const gyrCalBatchSize = 50
const gyrCalThreshold = 1_000_000
const gyrCalStableThreshold = 100_000

type GyrCal struct {
	stable bool
	offset [3]int32
	last   [3]int32
	sum    [3]int32
	count  [3]int32
}

func (g *GyrCal) get(x, y, z int32) (int32, int32, int32) {
	return x - g.offset[0], y - g.offset[1], z - g.offset[2]
}

func (g *GyrCal) apply(x, y, z int32) {
	g.applyAxis(0, x)
	g.applyAxis(1, y)
	g.applyAxis(2, z)
	if g.stable {
		return
	}
	g.stable = true
	for i := 0; i < 3; i++ {
		if g.last[i] == 0 || abs(g.last[i]) > gyrCalStableThreshold {
			g.stable = false
		}
	}
}

func (g *GyrCal) applyAxis(i, v int32) {
	tmp := v - g.offset[i]
	if tmp > 0 && tmp > gyrCalThreshold {
		tmp = gyrCalThreshold
	}
	if tmp < 0 && tmp < -gyrCalThreshold {
		tmp = -gyrCalThreshold
	}
	g.sum[i] += tmp
	g.count[i]++
	if g.count[i] > gyrCalBatchSize {
		g.last[i] = g.sum[i] / g.count[i] / 2
		g.offset[i] += g.last[i]
		g.count[i] = 0
		g.sum[i] = 0
	}
}

func abs(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
