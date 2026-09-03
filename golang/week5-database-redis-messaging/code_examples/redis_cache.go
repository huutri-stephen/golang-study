package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// Simulated Redis client interface (in production use github.com/redis/go-redis)
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
}

var ErrCacheMiss = errors.New("cache miss")

// --- 1. Cache-Aside Pattern ---

type UserCache struct {
	redis RedisClient
	db    UserDB
	ttl   time.Duration
}

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserDB interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	Update(ctx context.Context, user *User) error
}

func (c *UserCache) GetUser(ctx context.Context, id int64) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	// 1. Try cache
	data, err := c.redis.Get(ctx, key)
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(data), &user); err == nil {
			return &user, nil // cache hit
		}
	}

	// 2. Cache miss → query DB
	user, err := c.db.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Write to cache (with jitter to prevent avalanche)
	ttl := c.ttl + time.Duration(rand.Intn(60))*time.Second
	jsonData, _ := json.Marshal(user)
	c.redis.Set(ctx, key, string(jsonData), ttl)

	return user, nil
}

func (c *UserCache) UpdateUser(ctx context.Context, user *User) error {
	// 1. Update DB first
	if err := c.db.Update(ctx, user); err != nil {
		return err
	}

	// 2. Invalidate cache (NOT update!)
	key := fmt.Sprintf("user:%d", user.ID)
	c.redis.Del(ctx, key)

	return nil
}

// --- 2. Singleflight / Cache Stampede Prevention ---

type SingleFlightCache struct {
	redis    RedisClient
	db       UserDB
	mu       sync.Mutex
	inflight map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val *User
	err error
}

func (c *SingleFlightCache) GetUser(ctx context.Context, id int64) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	// Try cache first
	data, err := c.redis.Get(ctx, key)
	if err == nil {
		var user User
		json.Unmarshal([]byte(data), &user)
		return &user, nil
	}

	// Singleflight: only one goroutine loads from DB
	c.mu.Lock()
	if c.inflight == nil {
		c.inflight = make(map[string]*call)
	}

	if existing, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		existing.wg.Wait() // wait for in-flight request
		return existing.val, existing.err
	}

	// First request for this key
	cl := &call{}
	cl.wg.Add(1)
	c.inflight[key] = cl
	c.mu.Unlock()

	// Load from DB
	cl.val, cl.err = c.db.GetByID(ctx, id)
	if cl.err == nil {
		jsonData, _ := json.Marshal(cl.val)
		c.redis.Set(ctx, key, string(jsonData), 5*time.Minute)
	}

	// Signal waiting goroutines
	cl.wg.Done()

	// Cleanup
	c.mu.Lock()
	delete(c.inflight, key)
	c.mu.Unlock()

	return cl.val, cl.err
}

// --- 3. Distributed Lock ---

type DistributedLock struct {
	redis RedisClient
	key   string
	value string // unique identifier
	ttl   time.Duration
}

func NewDistributedLock(redis RedisClient, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		redis: redis,
		key:   "lock:" + key,
		value: fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63()),
		ttl:   ttl,
	}
}

func (l *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	// SET NX PX — atomic set-if-not-exists with expiry
	acquired, err := l.redis.SetNX(ctx, l.key, l.value, l.ttl)
	return acquired, err
}

func (l *DistributedLock) Release(ctx context.Context) error {
	// Must check value before delete (Lua script in production)
	// Only release if we own the lock
	val, err := l.redis.Get(ctx, l.key)
	if err != nil {
		return err
	}
	if val != l.value {
		return errors.New("lock owned by another process")
	}
	return l.redis.Del(ctx, l.key)
}

// Usage pattern:
func processWithLock(ctx context.Context, redis RedisClient, orderID string) error {
	lock := NewDistributedLock(redis, "order:"+orderID, 30*time.Second)

	acquired, err := lock.Acquire(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("could not acquire lock")
	}
	defer lock.Release(ctx)

	// Process order (exclusive access)
	fmt.Printf("Processing order %s exclusively\n", orderID)
	return nil
}

// --- 4. Cache with Null Value (Penetration Prevention) ---

const nullValue = "__NULL__"

type PenetrationSafeCache struct {
	redis RedisClient
	db    UserDB
}

func (c *PenetrationSafeCache) GetUser(ctx context.Context, id int64) (*User, error) {
	key := fmt.Sprintf("user:%d", id)

	data, err := c.redis.Get(ctx, key)
	if err == nil {
		if data == nullValue {
			return nil, ErrCacheMiss // cached null — don't hit DB
		}
		var user User
		json.Unmarshal([]byte(data), &user)
		return &user, nil
	}

	// DB lookup
	user, err := c.db.GetByID(ctx, id)
	if err != nil {
		// Cache null value with short TTL (prevent penetration)
		c.redis.Set(ctx, key, nullValue, 1*time.Minute)
		return nil, err
	}

	// Cache real value with longer TTL
	jsonData, _ := json.Marshal(user)
	c.redis.Set(ctx, key, string(jsonData), 10*time.Minute)
	return user, nil
}

func main() {
	log.Println("Redis cache patterns demo")
	fmt.Println(`
Patterns demonstrated:
1. Cache-Aside (read: cache → DB, write: DB → invalidate)
2. Singleflight (prevent cache stampede)
3. Distributed Lock (exclusive access)
4. Null Value Cache (prevent cache penetration)

Cache Strategy Decision:
• Read-heavy, tolerance for stale → Cache-Aside
• Need strong consistency → Read/Write-Through
• Write-heavy, async → Write-Behind
• Critical section → Distributed Lock

TTL Strategy:
• Base TTL + random jitter (prevent avalanche)
• Short TTL for null values (prevent penetration)
• Longer TTL for stable data
• Never expire for truly hot keys (background refresh)
`)
}
