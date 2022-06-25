package api

type Sensor interface {
	Check() (float64, interface{}, error)
}
