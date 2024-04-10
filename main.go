package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Cache represents a replicated memory cache.
type Cache struct {
	data  sync.Map   // Concurrent map to store data
	peers []string   // List of peer servers
	mu    sync.Mutex // Mutex for synchronizing access to the cache
}

// Add adds a new key-value pair to the cache and replicates it to peer servers.
func (c *Cache) Add(server string, key, value string) {
	c.mu.Lock() // Lock to ensure atomicity
	defer c.mu.Unlock()

	c.data.Store(key, value)            // Store data in cache
	c.replicateData(server, key, value) // Replicate data to peer servers
}

// Get retrieves the value associated with the given key from the cache.
func (c *Cache) Get(server string, key string) (string, bool) {
	val, ok := c.data.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// Delete removes the key-value pair associated with the given key from the cache.
func (c *Cache) Delete(server string, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data.Delete(key)             // Delete data from cache
	c.replicateDelete(server, key) // Replicate deletion to peer servers
}

// replicateData replicates the newly added data to all peer servers.
func (c *Cache) replicateData(server string, key, value string) {
	for _, peer := range c.peers {
		go func(peerAddr string) {
			_, err := http.Post(peerAddr+"/replicate", "application/json", bytes.NewBuffer([]byte(key+"="+value)))
			if err != nil {
				log.Printf("\033[31mError replicating data to %s: %v\033[0m", peerAddr, err)
			} else {
				log.Printf("\033[32mReplicated data to %s\033[0m", peerAddr)
			}
		}(peer)
	}
}

// replicateDelete replicates the deletion of data to all peer servers.
func (c *Cache) replicateDelete(server string, key string) {
	for _, peer := range c.peers {
		go func(peerAddr string) {
			req, err := http.NewRequest("DELETE", peerAddr+"/replicate/"+key, nil)
			if err != nil {
				log.Printf("\033[31mError creating delete request to %s: %v\033[0m", peerAddr, err)
				return
			}
			_, err = http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("\033[31mError replicating delete to %s: %v\033[0m", peerAddr, err)
			} else {
				log.Printf("\033[33mReplicated delete to %s\033[0m", peerAddr)
			}
		}(peer)
	}
}

// getDataFromCache retrieves all key-value pairs from the cache and formats them as a string.
func (c *Cache) getDataFromCache(server string) string {
	var buf bytes.Buffer
	c.data.Range(func(key, value interface{}) bool {
		buf.WriteString(fmt.Sprintf("%s=%s\n", key, value)) // Format key-value pair
		return true
	})
	log.Printf("\033[36mRetrieved cache content on %s\033[0m", server)
	return buf.String()
}

func main() {
	cache := &Cache{
		peers: []string{"http://localhost:8081", "http://localhost:8082"}, // Example peer addresses
	}

	// Handler for adding data to the cache
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if key == "" || value == "" {
			http.Error(w, "key and value are required", http.StatusBadRequest)
			return
		}
		cache.Add("http://localhost:8080", key, value) // Adding data to the cache
		log.Printf("\033[32mAdded key-value pair: %s=%s on http://localhost:8080\033[0m", key, value)
		w.WriteHeader(http.StatusOK)
	})

	// Handler for retrieving data from the cache
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}
		val, ok := cache.Get("http://localhost:8080", key)
		if !ok {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		log.Printf("\033[36mRetrieved value for key %s: %s on http://localhost:8080\033[0m", key, val)
		json.NewEncoder(w).Encode(map[string]string{key: val})
	})

	// Handler for deleting data from the cache
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}
		cache.Delete("http://localhost:8080", key)
		log.Printf("\033[33mDeleted key %s on http://localhost:8080\033[0m", key)
		w.WriteHeader(http.StatusOK)
	})

	// Handler for replicating data from other servers
	http.HandleFunc("/replicate", func(w http.ResponseWriter, r *http.Request) {
		var data map[string]string
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		for key, value := range data {
			cache.Add(r.Host, key, value) // Add replicated data to cache
		}
		log.Printf("\033[35mReplicated data from peer server on %s\033[0m", r.Host)
		w.WriteHeader(http.StatusOK)
	})

	// Handler for retrieving cache content
	http.HandleFunc("/getCacheContent", func(w http.ResponseWriter, r *http.Request) {
		content := cache.getDataFromCache(r.Host)
		fmt.Fprintf(w, "%s", content)
	})

	// Handler for SSE updates
	http.HandleFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Create a channel to send updates to clients
		updates := make(chan string)

		// Start a goroutine to listen for cache updates and send them to clients
		go func() {
			for {
				// Retrieve data from the cache and send it to clients
				data := cache.getDataFromCache(r.Host)
				updates <- data

				// Wait for a short duration before sending the next update
				time.Sleep(1 * time.Second)
			}
		}()

		// Continuously send updates to clients
		for {
			select {
			case update := <-updates:
				fmt.Fprintf(w, "data: %s\n\n", update) // Send update as SSE message
				w.(http.Flusher).Flush()               // Flush the response writer to ensure the message is sent immediately
			case <-r.Context().Done():
				return
			}
		}
	})

	// ListenAndServe on multiple ports
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	go func() {
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()
	go func() {
		log.Fatal(http.ListenAndServe(":8082", nil))
	}()

	select {}
}
