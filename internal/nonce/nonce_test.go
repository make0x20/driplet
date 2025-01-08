package nonce

import (
    "testing"
    "time"
    "sync"
	"fmt"
)

func TestMemoryStore(t *testing.T) {
    // Test basic store and check
    t.Run("basic store and check", func(t *testing.T) {
        store := NewMemoryStore()
        now := time.Now()
        future := now.Add(time.Minute)

        store.Store("nonce1", "endpoint1", future)
        
        if !store.Check("nonce1", "endpoint1") {
            t.Error("expected valid nonce to be found")
        }
        
        // Wrong endpoint
        if store.Check("nonce1", "endpoint2") {
            t.Error("expected check with wrong endpoint to fail")
        }
        
        // Non-existent nonce
        if store.Check("nonce2", "endpoint1") {
            t.Error("expected non-existent nonce check to fail")
        }
    })

    // Test expiration
    t.Run("expiration", func(t *testing.T) {
        store := NewMemoryStore()
        past := time.Now().Add(-time.Minute)
        
        store.Store("nonce1", "endpoint1", past)
        
        if store.Check("nonce1", "endpoint1") {
            t.Error("expected expired nonce check to fail")
        }
    })

    // Test concurrent access
    t.Run("concurrent access", func(t *testing.T) {
        store := NewMemoryStore()
        future := time.Now().Add(time.Minute)
        
        var wg sync.WaitGroup
        workers := 10
        iterations := 100
        
        // Concurrent writers
        for i := 0; i < workers; i++ {
            wg.Add(1)
            go func(workerID int) {
                defer wg.Done()
                for j := 0; j < iterations; j++ {
                    nonce := fmt.Sprintf("nonce-%d-%d", workerID, j)
                    store.Store(nonce, "endpoint1", future)
                }
            }(i)
        }
        
        // Concurrent readers
        for i := 0; i < workers; i++ {
            wg.Add(1)
            go func(workerID int) {
                defer wg.Done()
                for j := 0; j < iterations; j++ {
                    nonce := fmt.Sprintf("nonce-%d-%d", workerID, j)
                    store.Check(nonce, "endpoint1")
                }
            }(i)
        }
        
        wg.Wait()
    })

    // Test overwrite behavior
    t.Run("overwrite", func(t *testing.T) {
        store := NewMemoryStore()
        now := time.Now()
        future := now.Add(time.Minute)
        
        // Store initial value
        store.Store("nonce1", "endpoint1", future)
        
        // Overwrite with new endpoint
        store.Store("nonce1", "endpoint2", future)
        
        if store.Check("nonce1", "endpoint1") {
            t.Error("expected original endpoint check to fail after overwrite")
        }
        if !store.Check("nonce1", "endpoint2") {
            t.Error("expected new endpoint check to succeed after overwrite")
        }
    })

    // Test cleanup of expired entries during check
    t.Run("cleanup on check", func(t *testing.T) {
        store := NewMemoryStore()
        past := time.Now().Add(-time.Minute)
        
        store.Store("nonce1", "endpoint1", past)
        
        // First check should clean up the expired entry
        if store.Check("nonce1", "endpoint1") {
            t.Error("expected expired nonce check to fail")
        }
        
        // Verify entry was removed
        store.mu.RLock()
        _, exists := store.store["nonce1"]
        store.mu.RUnlock()
        
        if exists {
            t.Error("expected expired entry to be removed")
        }
    })
}
