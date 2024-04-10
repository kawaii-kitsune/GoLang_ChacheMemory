package main

import (
	"sync"
	"testing"
)

func TestAdd(t *testing.T) {
	// Set up a test cache instance
	cache := &Cache{
		data:  sync.Map{},
		peers: []string{"http://localhost:8081", "http://localhost:8082"},
	}

	// Call the Add function with test data
	key := "testKey"
	value := "testValue"
	cache.Add("http://localhost:8080", key, value)

	// Retrieve the value using Get function
	val, ok := cache.Get("http://localhost:8080", key)

	// Check if the value was added successfully
	if !ok || val != value {
		t.Errorf("Add function failed. Expected value: %s, Actual value: %s", value, val)
	}
}

func TestGet(t *testing.T) {
	// Set up a test cache instance with some initial data
	cache := &Cache{
		data:  sync.Map{},
		peers: []string{"http://localhost:8081", "http://localhost:8082"},
	}

	// Add test data to the cache
	cache.data.Store("testKey", "testValue")

	// Call the Get function to retrieve the value
	key := "testKey"
	expectedValue := "testValue"
	actualValue, ok := cache.Get("http://localhost:8080", key)

	// Check if the value was retrieved successfully
	if !ok || actualValue != expectedValue {
		t.Errorf("Get function failed. Expected value: %s, Actual value: %s", expectedValue, actualValue)
	}
}

func TestDelete(t *testing.T) {
	// Set up a test cache instance with some initial data
	cache := &Cache{
		data:  sync.Map{},
		peers: []string{"http://localhost:8081", "http://localhost:8082"},
	}

	// Add test data to the cache
	cache.data.Store("testKey", "testValue")

	// Call the Delete function to remove the data
	key := "testKey"
	cache.Delete("http://localhost:8080", key)

	// Attempt to retrieve the deleted value
	_, ok := cache.Get("http://localhost:8080", key)

	// Check if the value was deleted successfully
	if ok {
		t.Errorf("Delete function failed. Value still exists for key: %s", key)
	}
}
