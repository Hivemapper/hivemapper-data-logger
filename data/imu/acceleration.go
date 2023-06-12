package imu

type Orientation string

const (
	OrientationUnset Orientation = "Unset"
	OrientationFront Orientation = "OrientationFront"
	OrientationRight Orientation = "OrientationRight"
	OrientationLeft  Orientation = "OrientationLeft"
	OrientationBack  Orientation = "OrientationBack"
)

type AccelerationData interface {
	GetX() float64
	GetY() float64
	GetZ() float64
	GetMagnitude() float64
}

type OrientedAccelerationData interface {
	AccelerationData
	GetOrientation() Orientation
}

type TiltCorrectedAccelerationData interface {
	AccelerationData
	OrientedAccelerationData
	GetXAngle() float64
	GetYAngle() float64
}

type BaseAcceleration struct {
	x float64
	y float64
	z float64
	m float64
}

// NewBaseAcceleration The X, Y and Z are the values which come out of the RawImuEvent
func NewBaseAcceleration(x, y, z, m float64) *BaseAcceleration {
	return &BaseAcceleration{
		x: x,
		y: y,
		z: z,
		m: m,
	}
}

func (a *BaseAcceleration) GetX() float64 {
	return a.x
}

func (a *BaseAcceleration) GetY() float64 {
	return a.y
}

func (a *BaseAcceleration) GetZ() float64 {
	return a.z
}

func (a *BaseAcceleration) GetMagnitude() float64 {
	return a.m
}

type OrientedAcceleration struct {
	*BaseAcceleration
	xAngle      float64
	yAngle      float64
	zAngle      float64
	orientation Orientation
}

func NewOrientedAcceleration(acceleration *BaseAcceleration, orientation Orientation) *OrientedAcceleration {
	return &OrientedAcceleration{
		BaseAcceleration: acceleration,
		orientation:      orientation,
	}
}

func (o *OrientedAcceleration) GetOrientation() Orientation {
	return o.orientation
}

func (o *OrientedAcceleration) GetXAngle() float64 {
	return o.xAngle
}

func (o *OrientedAcceleration) GetYAngle() float64 {
	return o.yAngle
}

func (o *OrientedAcceleration) GetZAngle() float64 {
	return o.zAngle
}

func (o *OrientedAcceleration) SetXAngle(xAngle float64) {
	o.xAngle = xAngle
}

func (o *OrientedAcceleration) SetYAngle(yAngle float64) {
	o.yAngle = yAngle
}

func (o *OrientedAcceleration) SetZAngle(zAngle float64) {
	o.zAngle = zAngle
}
