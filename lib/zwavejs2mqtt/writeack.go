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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

const WriteValueCommandTopic = "/writeValue"

// zwave-js SetValueStatus values that mean the write was accepted
const (
	setValueStatusWorking             = 1
	setValueStatusSuccessUnsupervised = 254
	setValueStatusSuccess             = 255
)

// writeResponse is the response of zwave-js-ui to an api call, published on the request topic without "/set".
// Origin echoes the request payload, which is how requestId correlates the response.
type writeResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
	Origin  json.RawMessage `json:"origin"`
}

// parseWriteResponse returns the request id of the response and the evaluation of its outcome.
// ok is false if the payload is no response to a request carrying a requestId.
func parseWriteResponse(payload []byte) (requestId string, result error, ok bool) {
	resp := writeResponse{}
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", nil, false
	}
	origin := struct {
		RequestId string `json:"requestId"`
	}{}
	if len(resp.Origin) == 0 || json.Unmarshal(resp.Origin, &origin) != nil || origin.RequestId == "" {
		return "", nil, false
	}
	return origin.RequestId, evaluateWriteResponse(resp), true
}

func evaluateWriteResponse(resp writeResponse) error {
	if !resp.Success {
		if resp.Message == "" {
			return errors.New("zwave-js-ui reported failure")
		}
		return errors.New("zwave-js-ui reported failure: " + resp.Message)
	}
	status := struct {
		Status  *float64 `json:"status"`
		Message string   `json:"message"`
	}{}
	if len(resp.Result) == 0 || json.Unmarshal(resp.Result, &status) != nil || status.Status == nil {
		// no status to evaluate
		return nil
	}
	switch *status.Status {
	case setValueStatusWorking, setValueStatusSuccessUnsupervised, setValueStatusSuccess:
		return nil
	}
	if status.Message != "" {
		return fmt.Errorf("zwave-js-ui rejected write with status %v: %v", *status.Status, status.Message)
	}
	return fmt.Errorf("zwave-js-ui rejected write with status %v", *status.Status)
}

// pendingWrites correlates writeValue responses with waiting callers by request id
type pendingWrites struct {
	mux     sync.Mutex
	pending map[string]chan error
}

func newPendingWrites() *pendingWrites {
	return &pendingWrites{pending: map[string]chan error{}}
}

// register must be called before the request is published
func (this *pendingWrites) register(id string) chan error {
	ch := make(chan error, 1)
	this.mux.Lock()
	defer this.mux.Unlock()
	this.pending[id] = ch
	return ch
}

func (this *pendingWrites) remove(id string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	delete(this.pending, id)
}

// resolve completes the pending entry once; unknown ids and repeated responses are ignored
func (this *pendingWrites) resolve(id string, result error) bool {
	this.mux.Lock()
	ch, ok := this.pending[id]
	delete(this.pending, id)
	this.mux.Unlock()
	if !ok {
		return false
	}
	ch <- result // buffered with capacity 1 and removed from the map, so this never blocks
	return true
}

func (this *pendingWrites) len() int {
	this.mux.Lock()
	defer this.mux.Unlock()
	return len(this.pending)
}

func (this *pendingWrites) wait(ctx context.Context, id string, ch chan error, timeout time.Duration) error {
	defer this.remove(id)
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-ch:
		return err
	case <-timer.C:
		return fmt.Errorf("no response from zwave-js-ui within %v", timeout)
	case <-ctx.Done():
		return fmt.Errorf("waiting for response from zwave-js-ui aborted: %w", ctx.Err())
	}
}

func newRequestId() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// startWriteValueResponseListener is called on every (re)connect
func (this *Client) startWriteValueResponseListener() error {
	topic := this.apiTopic + WriteValueCommandTopic
	slog.Info("subscribe", "topic", topic)
	token := this.mqtt.Subscribe(topic, 2, func(_ paho.Client, message paho.Message) {
		this.handleWriteValueResponse(message.Payload())
	})
	if token.Wait() && token.Error() != nil {
		slog.Error("Error on Subscribe", "topic", topic, "error", token.Error())
		this.ForwardError("Error on Subscribe: " + token.Error().Error())
		return token.Error()
	}
	return nil
}

func (this *Client) handleWriteValueResponse(payload []byte) {
	id, result, ok := parseWriteResponse(payload)
	if !ok {
		return
	}
	this.writes.resolve(id, result)
}

// SetValueByValueIdWithAck publishes the write with a request id and waits until zwave-js-ui answered it.
// A timeout <= 0 disables the wait.
func (this *Client) SetValueByValueIdWithAck(ctx context.Context, valueId string, value interface{}, timeout time.Duration) error {
	if timeout <= 0 {
		return this.SetValueByValueId(valueId, value)
	}
	valueIdObject, err := parseValueId(valueId)
	if err != nil {
		return err
	}
	requestId, err := newRequestId()
	if err != nil {
		return err
	}
	ch := this.writes.register(requestId) // before publishing, the response may arrive immediately
	defer this.writes.remove(requestId)
	err = this.sendZwayCommand(WriteValueCommandTopic, []interface{}{valueIdObject, value}, map[string]interface{}{"requestId": requestId})
	if err != nil {
		return err
	}
	return this.writes.wait(ctx, requestId, ch, timeout)
}
