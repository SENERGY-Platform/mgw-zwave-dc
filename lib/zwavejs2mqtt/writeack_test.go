/*
 * Copyright (c) 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zwavejs2mqtt

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseWriteResponse(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		wantOk    bool
		wantId    string
		wantErr   bool
		errSubstr string
	}{
		{"status working", `{"success":true,"message":"Success zwave api call","result":{"status":1},"args":[],"origin":{"args":[],"requestId":"r1"}}`, true, "r1", false, ""},
		{"status success unsupervised", `{"success":true,"result":{"status":254},"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"status success", `{"success":true,"result":{"status":255},"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"status no device support", `{"success":true,"result":{"status":0},"origin":{"requestId":"r1"}}`, true, "r1", true, "status 0"},
		{"status fail", `{"success":true,"result":{"status":2,"message":"node dead"},"origin":{"requestId":"r1"}}`, true, "r1", true, "node dead"},
		{"status endpoint not found", `{"success":true,"result":{"status":3},"origin":{"requestId":"r1"}}`, true, "r1", true, "status 3"},
		{"status not implemented", `{"success":true,"result":{"status":4},"origin":{"requestId":"r1"}}`, true, "r1", true, "status 4"},
		{"status invalid value", `{"success":true,"result":{"status":5,"message":"bad"},"origin":{"requestId":"r1"}}`, true, "r1", true, "status 5"},
		{"success false", `{"success":false,"message":"Node 3 not found","origin":{"requestId":"r1"}}`, true, "r1", true, "Node 3 not found"},
		{"success false without message", `{"success":false,"origin":{"requestId":"r1"}}`, true, "r1", true, "failure"},
		{"missing status", `{"success":true,"result":{},"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"null result", `{"success":true,"result":null,"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"no result", `{"success":true,"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"non object result", `{"success":true,"result":true,"origin":{"requestId":"r1"}}`, true, "r1", false, ""},
		{"no request id", `{"success":true,"result":{"status":255},"origin":{"args":[]}}`, false, "", false, ""},
		{"no origin", `{"success":true,"result":{"status":255}}`, false, "", false, ""},
		{"origin not object", `{"success":true,"origin":"x"}`, false, "", false, ""},
		{"not json", `nope`, false, "", false, ""},
		{"node info response", `{"success":true,"result":[{"id":2}],"args":[],"origin":{"args":[]}}`, false, "", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err, ok := parseWriteResponse([]byte(tt.payload))
			if ok != tt.wantOk || id != tt.wantId {
				t.Fatalf("got ok=%v id=%q, want ok=%v id=%q", ok, id, tt.wantOk, tt.wantId)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err=%v, want error=%v", err, tt.wantErr)
			}
			if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("error %q does not contain %q", err, tt.errSubstr)
			}
		})
	}
}

func TestPendingWritesCorrelation(t *testing.T) {
	p := newPendingWrites()
	ch1 := p.register("a")
	ch2 := p.register("b")

	if p.resolve("unknown", nil) {
		t.Error("unknown id must be ignored")
	}
	if p.len() != 2 {
		t.Errorf("pending = %d, want 2", p.len())
	}

	if !p.resolve("a", nil) {
		t.Fatal("a not resolved")
	}
	if p.resolve("a", context.Canceled) {
		t.Error("duplicate response must not resolve twice")
	}
	if err := <-ch1; err != nil {
		t.Errorf("a: duplicate response overwrote result: %v", err)
	}
	select {
	case <-ch2:
		t.Error("b resolved by response for a")
	default:
	}
	p.remove("b")
	if p.len() != 0 {
		t.Errorf("pending = %d, want 0", p.len())
	}
}

func TestPendingWritesWait(t *testing.T) {
	t.Run("response", func(t *testing.T) {
		p := newPendingWrites()
		ch := p.register("a")
		go p.resolve("a", nil)
		if err := p.wait(context.Background(), "a", ch, time.Second); err != nil {
			t.Error(err)
		}
		if p.len() != 0 {
			t.Error("entry leaked")
		}
	})
	t.Run("timeout", func(t *testing.T) {
		p := newPendingWrites()
		ch := p.register("a")
		err := p.wait(context.Background(), "a", ch, 20*time.Millisecond)
		if err == nil || !strings.Contains(err.Error(), "no response") {
			t.Errorf("got %v", err)
		}
		if p.len() != 0 {
			t.Error("entry leaked")
		}
	})
	t.Run("context cancelled", func(t *testing.T) {
		p := newPendingWrites()
		ch := p.register("a")
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()
		err := p.wait(ctx, "a", ch, time.Minute)
		if err == nil || !strings.Contains(err.Error(), "aborted") {
			t.Errorf("got %v", err)
		}
		if p.len() != 0 {
			t.Error("entry leaked")
		}
	})
	t.Run("late response after timeout", func(t *testing.T) {
		p := newPendingWrites()
		ch := p.register("a")
		_ = p.wait(context.Background(), "a", ch, 10*time.Millisecond)
		if p.resolve("a", nil) {
			t.Error("late response must be ignored")
		}
	})
}

func TestEvaluateErrorMessages(t *testing.T) {
	_, err, _ := parseWriteResponse([]byte(`{"success":true,"result":{"status":2,"message":"node dead"},"origin":{"requestId":"r"}}`))
	if err == nil || !strings.Contains(err.Error(), "2") || !strings.Contains(err.Error(), "node dead") {
		t.Errorf("error must name status and message: %v", err)
	}
}
