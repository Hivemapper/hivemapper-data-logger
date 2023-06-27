package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *GeoJsonHandler) magic(lon, lat, heading float64, now time.Time) (float64, float64, float64, *Coordinate, *Coordinate, error) {
	if heading == 0 {
		return lon, lat, heading, nil, nil, nil
	}
	point := NewCoordinate(lon, lat)
	query := bson.M{
		"coordinates": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lon, lat},
				},
				"$minDistance": 0,
				"$maxDistance": 500,
			},
		},
	}

	cursor, err := h.pointCollection.Find(context.Background(), query)
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("finding points: %w", err)
	}
	defer cursor.Close(context.Background())
	var coords []*Coordinate
	for cursor.Next(context.Background()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}

		c := NewCoordinate(
			result["coordinates"].(primitive.A)[0].(float64),
			result["coordinates"].(primitive.A)[1].(float64),
		)
		c.Way = result["wayID"].(int64)
		coords = append(coords, c)
	}
	closests := ClosestPoint(point, coords)
	if len(closests) == 0 {
		return lon, lat, heading, nil, nil, nil
	}

	closest, otherClosest := ClosestWay(closests, point, heading, coords)
	if otherClosest == nil {
		return lon, lat, heading, nil, nil, nil
	}
	//fmt.Println("otherClosest", otherClosest)

	corrected := Correct(point, closest, otherClosest)
	//fmt.Println("corrected", corrected)
	nLon, nLat, nHeading, err := h.Update(corrected.Lon, corrected.Lat, heading, now)
	//fmt.Println("filtered", nLon, nLat, nHeading)
	return nLon, nLat, nHeading, closest, otherClosest, nil
}
