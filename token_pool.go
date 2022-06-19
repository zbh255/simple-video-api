package main

import (
	"sync"
	"time"
)

var tokenPool = func() *TokenPool {
	pool := &TokenPool{}
	pool.tokenMap = make(map[string]interface{},1024)
	pool.tokenTimer = NewLittleHeap(1024)
	pool.init()
	return pool
}()

type TokenPool struct {
	mu         sync.RWMutex
	tokenMap   map[string]interface{}
	tokenTimer *LittleHeap
}

type Token struct {
	Token string
	Data  interface{}
}

func (t *TokenPool) AddToken(token string, data interface{}, liveTime time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	timeOut := time.Duration(time.Now().Unix()) + liveTime
	t.tokenMap[token] = Token{
		Token: token,
		Data:  data,
	}
	t.tokenTimer.Insert(Node{
		TimeOut: timeOut,
		Data: token,
	})
}

func (t *TokenPool) GetToken(token string) (Token,bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tmp, ok := t.tokenMap[token]
	if ok {
		return tmp.(Token),ok
	}
	return Token{}, false
}

func (t *TokenPool) DeleteToken(token string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tokenMap, token)
}

func (t *TokenPool) init() {
	go func() {
		for {
			time.Sleep(time.Millisecond * 100)
			st := time.Now().Unix()
		loop:
			t.mu.Lock()
			if t.tokenTimer.Size() <= 0 {
				t.mu.Unlock()
				continue
			}
			token := t.tokenTimer.Peek()
			if token.TimeOut <= time.Duration(st) {
				tokenStr := t.tokenTimer.DelTop().Data.(string)
				delete(t.tokenMap, tokenStr)
				t.mu.Unlock()
				goto loop
			}
			t.mu.Unlock()
		}
	}()
}
