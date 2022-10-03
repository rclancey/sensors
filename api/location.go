package api

type Location struct {
	Country string `json:"country"`
	City string `json:"city"`
	State string `json:"state"`
	ZipCode string `json:"zipcode"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Elevation float64 `json:"elevation"`
}
