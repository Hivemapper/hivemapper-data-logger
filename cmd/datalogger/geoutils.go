package main

import (
	"math"
)

const (
	earthRadius = 6371 * 1000 // meters
)

type Coordinate struct {
	Lon float64
	Lat float64
	Way int64
}

func NewCoordinate(lon float64, lat float64) *Coordinate {
	return &Coordinate{
		Lon: lon,
		Lat: lat,
	}
}

func (c *Coordinate) HeadingTo(to *Coordinate) float64 {
	lat1 := degreesToRadians(c.Lat)
	lon1 := degreesToRadians(c.Lon)
	lat2 := degreesToRadians(to.Lat)
	lon2 := degreesToRadians(to.Lon)

	// Compute differences in longitude and latitude
	deltaLon := lon2 - lon1

	// Compute heading
	y := math.Sin(deltaLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)
	heading := math.Atan2(y, x)

	// Convert heading to degrees
	heading = heading * (180.0 / math.Pi)

	// Adjust heading to range from 0 to 360 degrees
	if heading < 0 {
		heading += 360.0
	}

	return heading
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func radiansToDegrees(r float64) float64 {
	return r * 180 / math.Pi
}

func Distance(c1, c2 *Coordinate) float64 {
	lat1 := degreesToRadians(c1.Lat)
	lon1 := degreesToRadians(c1.Lon)
	lat2 := degreesToRadians(c2.Lat)
	lon2 := degreesToRadians(c2.Lon)

	diffLat := lat2 - lat1
	diffLon := lon2 - lon1

	a := math.Pow(math.Sin(diffLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*
		math.Pow(math.Sin(diffLon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return c * earthRadius
}

func ClosestPoint(c *Coordinate, points []*Coordinate) (closestPoints []*Coordinate) {
	var closestDistance float64
	for _, point := range points {
		distance := Distance(c, point)
		if len(closestPoints) == 0 || distance < closestDistance {
			closestPoints = []*Coordinate{point}
			closestDistance = distance
			continue
		}

		if distance == closestDistance {
			closestPoints = append(closestPoints, point)
		}
	}
	return closestPoints
}

func ClosestWay(closests []*Coordinate, origin *Coordinate, heading float64, points []*Coordinate) (closest *Coordinate, otherClosest *Coordinate) {
	closestWayPoints := make([][2]*Coordinate, len(closests))

	for i, c := range closests {
		smallestDistance := 0.0
		for _, point := range points {
			if point.Way != c.Way || (point.Lon == c.Lon && point.Lat == c.Lat) {
				continue
			}
			h := c.HeadingTo(point)
			delta := math.Abs(h - heading)
			//fmt.Println("delta:", delta)

			//-73.43955486902293, 45.57520659584327
			//-73.43980859833134, 45.57504515375223
			if (delta > 30 && delta < 0) || (delta < 165 && delta > 195) {
				continue
			}
			corrected := Correct(origin, c, point)
			closestToCorrectDistance := Distance(c, corrected)
			correctedToPointDistance := Distance(corrected, point)

			wayDistance := Distance(c, point)
			sum := closestToCorrectDistance + correctedToPointDistance
			distDelta := math.Abs(sum - wayDistance)
			if distDelta > 1 {
				continue
			}

			_, b, _ := calculateTriangleAngles(c, origin, corrected)
			distance := Distance(origin, point)
			if closestWayPoints[i][1] == nil || (b > 88 && b < 92 && distance < smallestDistance) {
				closestWayPoints[i][0] = c
				closestWayPoints[i][1] = point
				//break
				smallestDistance = distance
			}
		}
	}
	//fmt.Println("closest others:", closestWayPoints[0][1], closestWayPoints[1][1])
	smallestDistance := 0.0
	var closestWayPoint [2]*Coordinate
	for _, d := range closestWayPoints {
		if d[1] == nil {
			continue
		}
		distance := Distance(origin, d[1])
		if closestWayPoint[1] == nil || distance < smallestDistance {
			closestWayPoint = d
			smallestDistance = distance
		}
	}

	return closestWayPoint[0], closestWayPoint[1]
}

func Correct(c, w1, w2 *Coordinate) *Coordinate {
	directionX := w2.Lon - w1.Lon
	directionY := w2.Lat - w1.Lat

	distanceAC := ((c.Lon-w1.Lon)*directionX + (c.Lat-w1.Lat)*directionY) / (directionX*directionX + directionY*directionY)

	return &Coordinate{
		Lon: w1.Lon + distanceAC*directionX,
		Lat: w1.Lat + distanceAC*directionY,
	}
}

func calculateTriangleAngles(closest, origin, corrected *Coordinate) (float64, float64, float64) {
	sideA := Distance(origin, corrected)
	sideB := Distance(closest, corrected)
	sideC := Distance(closest, origin)

	angleA := math.Acos((sideB*sideB + sideC*sideC - sideA*sideA) / (2 * sideB * sideC))
	angleB := math.Acos((sideA*sideA + sideC*sideC - sideB*sideB) / (2 * sideA * sideC))
	angleC := math.Acos((sideA*sideA + sideB*sideB - sideC*sideC) / (2 * sideA * sideB))

	angleADeg := angleA * (180 / math.Pi)
	angleBDeg := angleB * (180 / math.Pi)
	angleCDeg := angleC * (180 / math.Pi)
	return angleADeg, angleBDeg, angleCDeg
}

//function calculateTriangleAngles(pointA, pointB, pointC) {
//const sideA = calculateDistance(pointB, pointC);
//const sideB = calculateDistance(pointA, pointC);
//const sideC = calculateDistance(pointA, pointB);
//
//const angleA = Math.acos((sideB * sideB + sideC * sideC - sideA * sideA) / (2 * sideB * sideC));
//const angleB = Math.acos((sideA * sideA + sideC * sideC - sideB * sideB) / (2 * sideA * sideC));
//const angleC = Math.acos((sideA * sideA + sideB * sideB - sideC * sideC) / (2 * sideA * sideB));
//
//const angleADeg = angleA * (180 / Math.PI);
//const angleBDeg = angleB * (180 / Math.PI);
//const angleCDeg = angleC * (180 / Math.PI);
//
//return [angleADeg, angleBDeg, angleCDeg];
//}
