package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type MetadataResponse struct {
	AppID           string `json:"app_id"`
	PodID           string `json:"pod_id"`
	Region          string `json:"region"`
	Zone            string `json:"zone"`
	PublicEndpoints string `json:"public_endpoints"`
	PodIP           string `json:"pod_ip"`
	HostIP          string `json:"host_ip"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := MetadataResponse{
		AppID:           os.Getenv("BUNNYNET_MC_APPID"),
		PodID:           os.Getenv("BUNNYNET_MC_PODID"),
		Region:          os.Getenv("BUNNYNET_MC_REGION"),
		Zone:            os.Getenv("BUNNYNET_MC_ZONE"),
		PublicEndpoints: os.Getenv("BUNNYNET_MC_PUBLIC_ENDPOINTS"),
		PodIP:           os.Getenv("BUNNYNET_MC_PODIP"),
		HostIP:          os.Getenv("BUNNYNET_MC_HOSTIP"),
	}
	json.NewEncoder(w).Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Go Bunny, Go!",
		"version": "1.0.0",
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/metadata", metadataHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
