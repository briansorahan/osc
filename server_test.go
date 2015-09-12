package osc

import (
	"sync"
	"testing"
	"time"
)

func TestAddMsgHandler(t *testing.T) {
	server := NewServer("localhost:6677")
	err := server.AddMsgHandler("/address/test", func(msg *Message) {})
	if err != nil {
		t.Error("Expected that OSC address '/address/test' is valid")
	}
}

func TestAddMsgHandlerWithInvalidAddress(t *testing.T) {
	server := NewServer("localhost:6677")
	err := server.AddMsgHandler("/address*/test", func(msg *Message) {})
	if err == nil {
		t.Error("Expected error with '/address*/test'")
	}
}

func TestServerMessageDispatching(t *testing.T) {
	finish := make(chan bool)
	start := make(chan bool)
	var done sync.WaitGroup
	done.Add(2)

	// Start the OSC server in a new go-routine
	go func() {
		server := NewServer("localhost:6677")
		err := server.AddMsgHandler("/address/test", func(msg *Message) {
			if len(msg.Arguments) != 1 {
				t.Error("Argument length should be 1 and is: " + string(len(msg.Arguments)))
			}

			if msg.Arguments[0].(int32) != 1122 {
				t.Error("Argument should be 1122 and is: " + string(msg.Arguments[0].(int32)))
			}

			// Stop OSC server
			server.Close()
			finish <- true
		})

		if err != nil {
			t.Error("Error adding message handler")
		}

		start <- true
		server.ListenAndDispatch()
	}()

	go func() {
		timeout := time.After(5 * time.Second)
		select {
		case <-timeout:
		case <-start:
			time.Sleep(500 * time.Millisecond)
			client := NewClient("localhost:6677")
			msg := NewMessage("/address/test")
			msg.Append(int32(1122))
			client.Send(msg)
		}

		done.Done()

		select {
		case <-timeout:
		case <-finish:
		}
		done.Done()
	}()

	done.Wait()
}

// FIXME
func TestServerMessageReceiving(t *testing.T) {
	finish := make(chan bool)
	start := make(chan bool)
	var done sync.WaitGroup
	done.Add(2)

	// Start the server in a go-routine
	go func() {
		server := NewServer("localhost:6677")
		server.Listen()

		// Start the client
		start <- true

		err := <-server.Listening

		if err != nil {
			t.Fatal(err)
		}

		for {
			packet, err := server.ReceivePacket()

			if err != nil {
				t.Fatal(err)
			}

			if packet != nil {
				msg := packet.(*Message)
				if msg.CountArguments() != 2 {
					t.Errorf("Argument length should be 2 and is: %d\n", msg.CountArguments())
				}

				if msg.Arguments[0].(int32) != 1122 {
					t.Error("Argument should be 1122 and is: " + string(msg.Arguments[0].(int32)))
				}

				if msg.Arguments[1].(int32) != 3344 {
					t.Error("Argument should be 3344 and is: " + string(msg.Arguments[1].(int32)))
				}

				server.Close()
				finish <- true
			}
		}
	}()

	go func() {
		timeout := time.After(5 * time.Second)
		select {
		case <-timeout:
		case <-start:
			client := NewClient("localhost:6677")
			msg := NewMessage("/address/test")
			msg.Append(int32(1122))
			msg.Append(int32(3344))
			time.Sleep(500 * time.Millisecond)
			client.Send(msg)
		}

		done.Done()

		select {
		case <-timeout:
		case <-finish:
		}
		done.Done()
	}()

	done.Wait()
}

func TestServerIsNotRunningAndGetsClosed(t *testing.T) {
	server := NewServer("127.0.0.1:8000")
	err := server.Close()
	if err == nil {
		t.Errorf("Expected error if the the server is not running and it gets closed")
	}
}
