package main

import (
	"encoding/json"
	"time"
)

const lockKey = "match_lock"

type Matchmaker struct{}

// NewMatchmaker initializes a new Matchmaker instance and begins matching users.
func NewMatchmaker() *Matchmaker {
	m := &Matchmaker{}
	go m.matchmakingLoop()
	return m
}

func acquireLock() {
	for {
		ok, err := rdb.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}
		if ok {
			return // Lock acquired successfully
		}
		time.Sleep(100 * time.Millisecond) // Wait before retrying to acquire the lock
	}
}

func releaseLock() {
	rdb.Del(ctx, lockKey) // Release the lock by deleting the key
}

// matchUsers continuously checks the queue for users to match.
func (m *Matchmaker) matchmakingLoop() {
	pubsub := rdb.Subscribe(ctx, "user_joined")
	ch := pubsub.Channel()
	defer pubsub.Close()
	acquireLock()
	defer releaseLock()
	for {
		for rdb.LLen(ctx, "queue").Val() < 2 {
			releaseLock()
			<-ch
			acquireLock()
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
