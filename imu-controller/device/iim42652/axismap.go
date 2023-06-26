package iim42652

type ImuAxis string

type AxisMap struct {
	CamX ImuAxis
	CamY ImuAxis
	CamZ ImuAxis
	InvX bool
	InvY bool
	InvZ bool
}

func NewAxisMap(x, y, z string) *AxisMap {
	return &AxisMap{
		CamX: ImuAxis(x),
		CamY: ImuAxis(y),
		CamZ: ImuAxis(z),
	}
}

func (am *AxisMap) SetInvertedAxes(invX, invY, invZ bool) {
	am.InvX = invX
	am.InvY = invY
	am.InvZ = invZ
}

func (am *AxisMap) X(acceleration *Acceleration) float64 {
	switch am.CamX {
	case "X":
		if am.InvX {
			return -acceleration.X
		}
		return acceleration.X
	case "Y":
		if am.InvX {
			return -acceleration.Y
		}
		return acceleration.Y
	case "Z":
		if am.InvX {
			return -acceleration.Z
		}
		return acceleration.Z
	default:
		panic("invalid axis")
	}
}

func (am *AxisMap) Y(acceleration *Acceleration) float64 {
	switch am.CamY {
	case "X":
		if am.InvY {
			return -acceleration.X
		}
		return acceleration.X
	case "Y":
		if am.InvY {
			return -acceleration.Y
		}
		return acceleration.Y
	case "Z":
		if am.InvY {
			return -acceleration.Z
		}
		return acceleration.Z
	default:
		panic("invalid axis")
	}
}

func (am *AxisMap) Z(acceleration *Acceleration) float64 {
	switch am.CamZ {
	case "X":
		if am.InvZ {
			return -acceleration.X
		}
		return acceleration.X
	case "Y":
		if am.InvZ {
			return -acceleration.Y
		}
		return acceleration.Y
	case "Z":
		if am.InvZ {
			return -acceleration.Z
		}
		return acceleration.Z
	default:
		panic("invalid axis")
	}
}
