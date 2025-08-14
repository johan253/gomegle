package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const lockKey = "match_lock"

// Lua: delete only if token matches
var luaUnlock = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
else
  return 0
end`)

// Lua: extend TTL only if token matches
var luaExtend = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("PEXPIRE", KEYS[1], ARGV[2])
else
  return 0
end`)

type Matchmaker struct {
	lockToken string // Token to identify the lock owner
}

// NewMatchmaker initializes a new Matchmaker instance and begins matching users.
func NewMatchmaker() *Matchmaker {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	m := &Matchmaker{
		lockToken: token,
	}
	go m.matchmakingLoop()
	return m
}

func (m *Matchmaker) acquireLock() {
	for {
		ok, err := rdb.SetNX(ctx, lockKey, m.lockToken, 5*time.Second).Result()
		if err != nil {
			panic(err)
		}
		if ok {
			return // Lock acquired successfully
		}
		time.Sleep(100 * time.Millisecond) // Wait before retrying to acquire the lock
	}
}

func (m *Matchmaker) releaseLock() {
	_, _ = luaUnlock.Run(ctx, rdb, []string{lockKey}, m.lockToken).Result()
}

// matchUsers continuously checks the queue for users to match.
func (m *Matchmaker) matchmakingLoop() {
	pubsub := rdb.Subscribe(ctx, "user_joined")
	ch := pubsub.Channel()
	defer pubsub.Close() //nolint:all
	m.acquireLock()
	defer m.releaseLock()
	for {
		for rdb.LLen(ctx, "queue").Val() < 2 {
			m.releaseLock()
			<-ch
			m.acquireLock()
		}
		u1, _ := rdb.LPop(ctx, "queue").Result()
		u2, _ := rdb.LPop(ctx, "queue").Result()

		joinMsg1 := ChatMsg{
			Type:    ChatMsgTypeJoin,
			Content: u2,
		}
		joinMsg2 := ChatMsg{
			Type:    ChatMsgTypeJoin,
			Content: u1,
		}
		data, _ := json.Marshal(joinMsg1)
		rdb.Publish(ctx, "user:"+u1, data)
		data, _ = json.Marshal(joinMsg2)
		rdb.Publish(ctx, "user:"+u2, data)

		rdb.SRem(ctx, "users", u1, u2) // Remove users from the active set
		_, _ = luaExtend.Run(ctx, rdb, []string{lockKey}, m.lockToken, int64(5000*time.Millisecond)).Result()
	}
}

// Enqueue adds a user to the matchmaker queue
func (m *Matchmaker) Enqueue(u *User) error {
	pipe := rdb.TxPipeline()
	pipe.RPush(ctx, "queue", u.pubKey)
	pipe.SAdd(ctx, "users", u.pubKey)
	pipe.Publish(ctx, "user_joined", "")
	_, err := pipe.Exec(ctx)
	return err
}

// Dequeue removes a user from the queue and closes their send channel. If the user is
// not found, this function is a no-op.
func (m *Matchmaker) Dequeue(u *User) error {
	pipe := rdb.TxPipeline()
	pipe.LRem(ctx, "queue", 0, u.pubKey)
	pipe.SRem(ctx, "users", u.pubKey)
	_, err := pipe.Exec(ctx)
	return err
}

// HasUser checks if a user with the given public key is currently in the matchmaker.
func (m *Matchmaker) HasUser(key string) (bool, error) {
	return rdb.SIsMember(ctx, "users", key).Result()
}
