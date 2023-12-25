package multiplex

import (
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

func NewMessageMux[P constraints.Ordered, Message any](getPattern func(Message) (P, error)) *MessageMux[P, Message] {
	return &MessageMux[P, Message]{
		getPattern:       getPattern,
		handlerByPattern: make(map[P]MessageFunc[Message]),
	}
}

type MessageMux[P constraints.Ordered, Message any] struct {
	mu sync.RWMutex

	// getPattern 是為了避免 generic type 呼叫 method 所造成的效能降低
	// 所以 Message type constraint 使用 any, 而沒有定義 GetPattern method
	//
	// https://www.youtube.com/watch?v=D1hI55EcBB4&t=20260s
	//
	// https://hackmd.io/@fieliapm/BkHvJjYq3#/5/2
	getPattern       func(Message) (P, error)
	handlerByPattern map[P]MessageFunc[Message]
	chainByGlobal    MessageChain[Message]

	notFoundHandler MessageFunc[Message]
}

func (mux *MessageMux[P, Message]) serve(dto Message) error {
	pattern, err := mux.getPattern(dto)
	if err != nil {
		return err
	}

	fn, ok := mux.handlerByPattern[pattern]
	if !ok {
		if mux.notFoundHandler == nil {
			return ErrNotFoundHandler
		}
		return mux.notFoundHandler(dto)
	}
	return mux.chainByGlobal.Link(fn)(dto)
}

// ServeWithoutMutex is a method of the MessageMux type that allows for serving a Message directly, without acquiring a read lock.
// It calls the serve method internally to handle the Message.
func (mux *MessageMux[P, Message]) ServeWithoutMutex(dto Message) error {
	return mux.serve(dto)
}

func (mux *MessageMux[P, Message]) Serve(dto Message) error {
	mux.mu.RLock()
	defer mux.mu.RUnlock()
	return mux.serve(dto)
}

func (mux *MessageMux[P, Message]) RegisterFunc(pattern P, fn MessageFunc[Message]) *MessageMux[P, Message] {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	_, ok := mux.handlerByPattern[pattern]
	if ok {
		panic(fmt.Sprintf("mux have duplicate pattern=%v", pattern))
	}

	mux.handlerByPattern[pattern] = fn
	return mux
}

func (mux *MessageMux[P, Message]) RemoveFunc(pattern P) {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	delete(mux.handlerByPattern, pattern)
}

func (mux *MessageMux[P, Message]) AddMiddleware(middlewares ...MessageDecorator[Message]) *MessageMux[P, Message] {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.chainByGlobal.AddChain(middlewares...)
	return mux
}

func (mux *MessageMux[P, Message]) SetNotFoundHandler(fn MessageFunc[Message]) *MessageMux[P, Message] {
	mux.mu.Lock()
	defer mux.mu.Unlock()
	mux.notFoundHandler = fn
	return mux
}
