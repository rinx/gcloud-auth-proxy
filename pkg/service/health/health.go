package health

import "net/http"

type Health interface {
	IsHealthy() bool
	IsReady() bool
}

type HealthCheck interface {
	Register(checkees ...Health)
	Healthz() http.Handler
	Readyz() http.Handler
}

type healthCheck struct {
	healthzMux *http.ServeMux
	readyzMux  *http.ServeMux

	checkees []Health
}

func New() HealthCheck {
	hc := &healthCheck{
		healthzMux: &http.ServeMux{},
		readyzMux:  &http.ServeMux{},
	}

	hc.healthzMux.HandleFunc("/", hc.healthz)
	hc.readyzMux.HandleFunc("/", hc.readyz)

	return hc
}

func (hc *healthCheck) Register(checkees ...Health) {
	hc.checkees = append(hc.checkees, checkees...)
}

func (hc *healthCheck) Healthz() http.Handler {
	return hc.healthzMux
}

func (hc *healthCheck) Readyz() http.Handler {
	return hc.readyzMux
}

func (hc *healthCheck) healthz(w http.ResponseWriter, r *http.Request) {
	for _, checkee := range hc.checkees {
		if !checkee.IsHealthy() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (hc *healthCheck) readyz(w http.ResponseWriter, r *http.Request) {
	for _, checkee := range hc.checkees {
		if !checkee.IsReady() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
