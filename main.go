package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
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

var (
	rdb *redis.Client
	ctx = context.Background()
)

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

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	err := rdb.Set(ctx, payload.Key, payload.Value, 0).Err()
	if err != nil {
		log.Printf("Redis SET error: %v", err)
		http.Error(w, "Failed to set value", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"key":    payload.Key,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Redis GET error: %v", err)
		http.Error(w, "Failed to get value", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"key":   key,
		"value": val,
	})
}

func main() {
	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local development
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // No password set
		DB:       0,  // Use default DB
	})

	// Test Redis connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis at %s: %v", redisAddr, err)
	} else {
		log.Printf("Successfully connected to Redis at %s", redisAddr)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/metadata", metadataHandler)
	mux.HandleFunc("/set", setHandler)
	mux.HandleFunc("/get", getHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
