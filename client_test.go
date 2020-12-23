package yass

import (
	"context"
	"errors"
	"testing"

	"github.com/michael-diggin/yass/mocks"
)

func TestClientPing(t *testing.T) {
	tt := []struct {
		name    string
		serving bool
	}{
		{"serving", true},
		{"not serving", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockgRPC := &mocks.MockCacheClient{}
			mockgRPC.PingFn = func() error {
				if tc.name == "serving" {
					return nil
				}
				return errors.New("Not serving")
			}
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			ok, _ := cc.Ping(context.Background())
			if ok != tc.serving {
				t.Fatalf("Serving status wrong")
			}
			if !mockgRPC.PingInvoked {
				t.Fatalf("Ping function never called")
			}
		})
	}
}

func TestClientSetValue(t *testing.T) {
	mockgRPC := &mocks.MockCacheClient{}
	mockgRPC.SetFn = func(ctx context.Context, key string, value []byte) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	val := "value"
	err := cc.SetValue(context.Background(), key, val)
	if err != nil {
		t.Fatalf("Non nil err: %v", err)
	}
	if !mockgRPC.SetInvoked {
		t.Fatalf("Add method was not invoked")
	}
}

var errTest = errors.New("Not in Cache")

func TestClientGetValue(t *testing.T) {
	tt := []struct {
		name  string
		key   string
		value string
		err   error
	}{
		{"valid case", "test", "value", nil},
		{"err case", "bad", "", errTest},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockgRPC := &mocks.MockCacheClient{}
			mockgRPC.GetFn = func(ctx context.Context, key string) ([]byte, error) {
				if key == "test" {
					return []byte(`"value"`), nil
				}
				return []byte{}, errTest
			}
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			val, err := cc.GetValue(context.Background(), tc.key)
			if err != tc.err {
				t.Fatalf("Unexpected err: %v", err)
			}
			if val != tc.value {
				t.Fatalf("Expected '%s', got %s", tc.value, val)
			}
			if !mockgRPC.GetInvoked {
				t.Fatalf("Get method was not invoked")
			}
		})
	}
}

func TestClientDelValue(t *testing.T) {
	mockgRPC := &mocks.MockCacheClient{}
	mockgRPC.DelFn = func(ctx context.Context, key string) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	err := cc.DelValue(context.TODO(), key)
	if err != nil {
		t.Fatalf("Non nil err: %v", err)
	}
	if !mockgRPC.DelInvoked {
		t.Fatalf("Del method was not invoked")
	}
}
