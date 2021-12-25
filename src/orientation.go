package main

import "math"

const radToDeg = 180 / math.Pi // 57.29578
const degToRad = 1 / radToDeg  // 0.0174533

// angles are in degrees
func quaternionToAngles(q [4]float64) (roll, pitch, yaw float32) {
	roll = float32(math.Atan2(q[0]*q[1]+q[2]*q[3], 0.5-q[1]*q[1]-q[2]*q[2]) * radToDeg)
	pitch = float32(math.Asin(-2.0*(q[1]*q[3]-q[0]*q[2])) * radToDeg)
	yaw = float32(math.Atan2(q[1]*q[2]+q[0]*q[3], 0.5-q[2]*q[2]-q[3]*q[3]) * radToDeg)
	return
}

func angleMinusAngle(a, b float32) float32 {
	diff := a - b
	if diff < -180 {
		diff += 360
	}
	if diff > 180 {
		diff -= 360
	}
	return diff
}

func angleToChannel(angle float32, max float32) uint16 {
	result := uint16(1500 + 500/max*angle)
	if result < 988 {
		return 988
	}
	if result > 2012 {
		return 2012
	}
	return result
}

// -- Quaternion initialisation --

// Initialize quaternion from current orientation (sensors)
// Finds North, then aligns North with (1, 0, 0) and Gravity with (0, 0, 1)
func orientationToQuaternion(ax, ay, az, mx, my, mz float64) (q [4]float64) {

	// Reset quaternion, we are searching from scratch
	q[0], q[1], q[2], q[3] = 1, 0, 0, 0

	// Find North
	wx, wy, wz := cross(ax, ay, az, mx, my, mz)
	nx, ny, nz := cross(wx, wy, wz, ax, ay, az)

	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", ax, ay, az)
	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", mx, my, mz)
	// fmt.Printf("%+0.3f %+0.3f %+0.3f\r\n", nx, ny, nz)

	// Find rotation of (nx, ny, nz) to align with (1, 0, 0)
	q = align(nx, ny, nz, 1, 0, 0, q)

	// Rotate (ax, ay, az) same amount
	ax, ay, az = rotate(ax, ay, az, q)

	// Find next rotation of (ax, ay, az) to align with (0, 0, 1)
	q = align(ax, ay, az, 0, 0, 1, q)
	return
}

// align two vectors and return respective rotation quaternion
func align(ax, ay, az, bx, by, bz float64, u [4]float64) (q [4]float64) {
	// float va, vx, vy, vz; // rotation angle and vector
	vx, vy, vz := cross(ax, ay, az, bx, by, bz)
	ax, ay, az = norm(ax, ay, az)
	bx, by, bz = norm(bx, by, bz)
	vx, vy, vz = norm(vx, vy, vz)
	va := math.Acos(dot(ax, ay, az, bx, by, bz))
	a2 := math.Cos(va / 2)
	b2 := vx * math.Sin(va/2)
	c2 := vy * math.Sin(va/2)
	d2 := vz * math.Sin(va/2)
	v := [4]float64{a2, b2, c2, d2}
	q = combine(u, v)
	return
}

// combine two rotations and return result rotation
// see https://en.wikipedia.org/wiki/Euler–Rodrigues_formula#Composition_of_rotations
func combine(u, v [4]float64) (q [4]float64) {
	a1, b1, c1, d1 := u[0], u[1], u[2], u[3]
	a2, b2, c2, d2 := v[0], v[1], v[2], v[3]
	q[0] = a1*a2 - b1*b2 - c1*c2 - d1*d2
	q[1] = a1*b2 + b1*a2 - c1*d2 + d1*c2
	q[2] = a1*c2 + c1*a2 - d1*b2 + b1*d2
	q[3] = a1*d2 + d1*a2 - b1*c2 + c1*b2
	return
}

// rotate given vector as defined by quaternion
// see https://en.wikipedia.org/wiki/Euler–Rodrigues_formula#Vector_formulation
func rotate(ax, ay, az float64, q [4]float64) (bx, by, bz float64) {
	r1x, r1y, r1z := cross(q[1], q[2], q[3], ax, ay, az)
	r2x, r2y, r2z := cross(q[1], q[2], q[3], r1x, r1y, r1z)
	bx = ax + 2*q[0]*r1x + 2*r2x
	by = ay + 2*q[0]*r1y + 2*r2y
	bz = az + 2*q[0]*r1z + 2*r2z
	return
}

// cross product of two vectors
func cross(ax, ay, az, bx, by, bz float64) (cx, cy, cz float64) {
	cx = ay*bz - az*by
	cy = az*bx - ax*bz
	cz = ax*by - ay*bx
	return
}

// dot product of two vectors
func dot(ax, ay, az, bx, by, bz float64) float64 {
	return ax*bx + ay*by + az*bz
}

// norm of a vector
func norm(ax, ay, az float64) (bx, by, bz float64) {
	len := math.Sqrt(dot(ax, ay, az, ax, ay, az))
	bx = ax / len
	by = ay / len
	bz = az / len
	return
}
