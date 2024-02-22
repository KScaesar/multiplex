package multiplex

import (
	"errors"
	"fmt"
	"time"
)

// https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers#pings_and_pongs_the_heartbeat_of_websockets
// https://stackoverflow.com/a/59273632/9288569

// https://stackoverflow.com/questions/49408031/websockets-in-chrome-and-firefox-disconnecting-after-one-minute-of-inactivity
// Most browsers trust the server to implement the ping, which is polite (since servers need to manage their load and their timeout for pinging...

// https://en.wikipedia.org/wiki/Comparison_of_WebSocket_implementations
// Automatic pongs for pings
// Automatic heartbeat pings

func WaitPingSendPong(ping <-chan error, pong func() error, isStop func() bool, pingWaitSecond int) error {
	pingWaitTime := time.Duration(pingWaitSecond) * time.Second

	timer := time.NewTimer(pingWaitTime)
	defer timer.Stop()

	for !isStop() {
		select {
		case <-timer.C:
			return errors.New("wait ping timeout")

		case err := <-ping:
			if err != nil {
				return fmt.Errorf("handle ping: %v", err)
			}

			err = pong()
			if err != nil {
				return fmt.Errorf("send pong: %v", err)
			}

			ok := timer.Reset(pingWaitTime)
			if !ok {
				timer = time.NewTimer(pingWaitTime)
			}
		}
	}

	return nil
}

func SendPingWaitPong(ping func() error, pong <-chan error, isStop func() bool, pongWaitSecond int) error {
	pongWaitTime := time.Duration(pongWaitSecond) * time.Second
	pingPeriod := pongWaitTime / 2

	done := make(chan struct{})
	defer close(done)

	notify := make(chan error, 2)

	sendPing := func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for !isStop() {
			select {
			case <-ticker.C:
				err := ping()
				if err != nil {
					notify <- fmt.Errorf("send ping: %v", err)
					return
				}

			case <-done:
				return
			}
		}
		notify <- nil
	}

	waitPong := func() {
		timer := time.NewTimer(pongWaitTime)
		defer timer.Stop()

		for !isStop() {
			select {
			case <-timer.C:
				notify <- errors.New("wait pong timeout")
				return

			case err := <-pong:
				if err != nil {
					notify <- fmt.Errorf("handle pong: %v", err)
					return
				}

				ok := timer.Reset(pongWaitTime)
				if !ok {
					timer = time.NewTimer(pongWaitTime)
				}

			case <-done:
				return
			}
		}
		notify <- nil
	}

	go sendPing()
	go waitPong()
	return <-notify
}
