package steam

import (
	"bytes"
	"testing"
	"time"

	"github.com/MeidoCompany/go-steam/v3/protocol"
	"github.com/MeidoCompany/go-steam/v3/protocol/protobuf"
	"github.com/MeidoCompany/go-steam/v3/protocol/steamlang"
	"google.golang.org/protobuf/proto"
)

func TestHandleLogOnResponseWithoutNonceDoesNotPanic(t *testing.T) {
	client := NewClient()

	body := &protobuf.CMsgClientLogonResponse{
		Eresult:          proto.Int32(int32(steamlang.EResult_OK)),
		HeartbeatSeconds: proto.Int32(1),
	}

	msg := protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientLogOnResponse, body)
	msg.SetSessionId(123)
	msg.SetSteamId(456)

	var buf bytes.Buffer
	if err := msg.Serialize(&buf); err != nil {
		t.Fatalf("serialize logon response: %v", err)
	}

	packet, err := protocol.NewPacket(buf.Bytes())
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		client.Auth.handleLogOnResponse(packet)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handleLogOnResponse timed out")
	}

	select {
	case event := <-client.Events():
		if _, ok := event.(*LoggedOnEvent); !ok {
			t.Fatalf("expected LoggedOnEvent, got %T", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected LoggedOnEvent")
	}

	if client.heartbeat != nil {
		client.heartbeat.Stop()
	}
}

func TestHandleLogOnResponseSkipsZeroHeartbeat(t *testing.T) {
	client := NewClient()

	body := &protobuf.CMsgClientLogonResponse{
		Eresult:          proto.Int32(int32(steamlang.EResult_OK)),
		HeartbeatSeconds: proto.Int32(0),
	}

	msg := protocol.NewClientMsgProtobuf(steamlang.EMsg_ClientLogOnResponse, body)
	msg.SetSessionId(123)
	msg.SetSteamId(456)

	var buf bytes.Buffer
	if err := msg.Serialize(&buf); err != nil {
		t.Fatalf("serialize logon response: %v", err)
	}

	packet, err := protocol.NewPacket(buf.Bytes())
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}

	client.Auth.handleLogOnResponse(packet)

	if client.heartbeat != nil {
		t.Fatal("expected zero heartbeat interval to skip ticker startup")
	}
}
