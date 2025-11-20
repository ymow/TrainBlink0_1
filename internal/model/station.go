package model

import (
	"time"
)

// StationType represents the type of station
type StationType string

const (
	StationTypeTRA  StationType = "TRA"  // Taiwan Railways
	StationTypeTHSR StationType = "THSR" // Taiwan High Speed Rail
	StationTypeMRT  StationType = "MRT"  // Metro
)

// Station represents a train station with geofencing data
type Station struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	Name      string      `json:"name" gorm:"not null"`
	NameEn    string      `json:"name_en" gorm:"column:name_en;not null"`
	Latitude  float64     `json:"lat" gorm:"column:lat;type:decimal(10,7);not null"`
	Longitude float64     `json:"lng" gorm:"column:lng;type:decimal(10,7);not null"`
	Type      StationType `json:"type" gorm:"type:varchar(10);not null"`
	Radius    int         `json:"radius" gorm:"default:500"` // meters
	Lines     []string    `json:"lines" gorm:"type:text[]"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// TableName overrides the table name
func (Station) TableName() string {
	return "stations"
}

// Coordinate represents GPS coordinates
type Coordinate struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// GetCoordinate returns the station's GPS coordinates
func (s *Station) GetCoordinate() Coordinate {
	return Coordinate{
		Latitude:  s.Latitude,
		Longitude: s.Longitude,
	}
}

// DisplayName returns the formatted display name
func (s *Station) DisplayName() string {
	return s.Name + " (" + string(s.Type) + ")"
}
