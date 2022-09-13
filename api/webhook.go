package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrInProgress = errors.New("webhook call already in progress")

type Webhook interface {
	Callback(body interface{}) error
}

type BaseWebhook struct {
	CallbackMethod string `json:"callback_method,omitempty"`
	CallbackURL string `json:"callback_url"`
	CallbackHeaders http.Header `json:"callback_headers,omitempty"`
	working bool
}

func (hook *BaseWebhook) Equals(other *BaseWebhook) bool {
	if hook.CallbackMethod != other.CallbackMethod {
		return false
	}
	if hook.CallbackURL != other.CallbackURL {
		return false
	}
	if hook.CallbackHeaders == nil || len(hook.CallbackHeaders) == 0 {
		if other.CallbackHeaders != nil && len(other.CallbackHeaders) != 0 {
			return false
		}
		return true
	}
	if other.CallbackHeaders == nil || len(other.CallbackHeaders) == 0 {
		return false
	}
	delim := "\n\n"
	for k, vs := range hook.CallbackHeaders {
		otherVs := other.CallbackHeaders.Values(k)
		if len(vs) != len(otherVs) {
			return false
		}
		if strings.Join(vs, delim) != strings.Join(otherVs, delim) {
			return false
		}
	}
	return true
}

func (hook *BaseWebhook) setParam(query url.Values, key string, val interface{}) error {
	switch v := val.(type) {
	case string:
		query.Add(key, v)
	case []byte:
		query.Add(key, string(v))
	case bool:
		query.Add(key, strconv.FormatBool(v))
	case int:
		query.Add(key, strconv.Itoa(v))
	case float64:
		query.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
	default:
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			n := rv.Len()
			for i := 0; i < n; i += 1 {
				err := hook.setParam(query, key, rv.Index(i).Interface())
				if err != nil {
					return err
				}
			}
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
			return hook.setParam(query, key, int(rv.Int()))
		case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
			return hook.setParam(query, key, int(rv.Uint()))
		case reflect.Map, reflect.Struct:
			data, err := json.Marshal(val)
			if err != nil {
				return err
			}
			return hook.setParam(query, key, data)
		case reflect.Ptr:
			if rv.IsNil() {
				return hook.setParam(query, key, "null")
			}
			return hook.setParam(query, key, rv.Elem().Interface())
		default:
			return fmt.Errorf("unhandled param type: %T", val)
		}
	}
	return nil
}

func (hook *BaseWebhook) Callback(body interface{}) error {
	if hook.working {
		return ErrInProgress
	}
	hook.working = true
	defer func() {
		hook.working = false
	}()

	u, err := url.Parse(hook.CallbackURL)
	if err != nil {
		return err
	}
	query := u.Query()
	headers := http.Header{}
	if hook.CallbackHeaders != nil {
		headers = hook.CallbackHeaders.Clone()
	}
	var reqBody io.ReadCloser
	if body != nil {
		switch hook.CallbackMethod {
		case http.MethodGet:
			bodyJson, err := json.Marshal(body)
			if err != nil {
				return err
			}
			bodyMap := map[string]interface{}{}
			err = json.Unmarshal(bodyJson, &bodyMap)
			if err != nil {
				return err
			}
			for k, iv := range bodyMap {
				err = hook.setParam(query, k, iv)
				if err != nil {
					return err
				}
			}
			u.RawQuery = query.Encode()
		case http.MethodPost, http.MethodPut:
			data, err := json.Marshal(body)
			if err != nil {
				return err
			}
			reqBody = ioutil.NopCloser(bytes.NewReader(data))
			headers.Set("Content-Type", "application/json")
		}
	}
	req, err := http.NewRequest(hook.CallbackMethod, u.String(), reqBody)
	if err != nil {
		return err
	}
	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status)
	}
	return nil
}

type ThresholdDirection string
const (
	ThresholdDirectionIncreasing = "increasing"
	ThresholdDirectionDecreasing = "decreasing"
)

type ThresholdWebhook struct {
	*BaseWebhook
	Direction ThresholdDirection `json:"direction"`
	TriggerThreshold float64 `json:"trigger_threshold"`
	ResetThreshold float64 `json:"reset_threshold"`
	triggered bool
}

func (hook *ThresholdWebhook) Evaluate(val float64, data interface{}) error {
	if hook.triggered {
		switch hook.Direction {
		case ThresholdDirectionDecreasing:
			if val >= hook.ResetThreshold {
				hook.triggered = false
			}
		case ThresholdDirectionIncreasing:
			if val <= hook.ResetThreshold {
				hook.triggered = false
			}
		}
	} else {
		switch hook.Direction {
		case ThresholdDirectionDecreasing:
			if val <= hook.TriggerThreshold {
				err := hook.Callback(data)
				if err != nil {
					return err
				}
				hook.triggered = true
			}
		case ThresholdDirectionIncreasing:
			if val >= hook.TriggerThreshold {
				err := hook.Callback(data)
				if err != nil {
					return err
				}
				hook.triggered = true
			}
		}
	}
	return nil
}

func (hook *ThresholdWebhook) Equals(other *ThresholdWebhook) bool {
	if !hook.BaseWebhook.Equals(other.BaseWebhook) {
		return false
	}
	if hook.Direction != other.Direction {
		return false
	}
	if hook.TriggerThreshold != other.TriggerThreshold {
		return false
	}
	if hook.ResetThreshold != other.ResetThreshold {
		return false
	}
	return true
}

type ThresholdWebhookList struct {
	fn string
	webhooks []*ThresholdWebhook
	stop chan bool
	lock *sync.Mutex
}

func NewThresholdWebhookList(fn string) (*ThresholdWebhookList, error) {
	webhooks := []*ThresholdWebhook{}
	err := ReadJSONFile(fn, &webhooks)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	return &ThresholdWebhookList{
		fn: fn,
		webhooks: webhooks,
		lock: &sync.Mutex{},
	}, nil
}

func (twl *ThresholdWebhookList) Save() error {
	data, err := json.Marshal(twl.webhooks)
	if err != nil {
		return err
	}
	f, err := os.Create(twl.fn)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (twl *ThresholdWebhookList) RegisterWebhook(webhook *ThresholdWebhook) {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	if webhook.BaseWebhook.CallbackMethod == "" {
		webhook.BaseWebhook.CallbackMethod = http.MethodGet
	}
	for _, hook := range twl.webhooks {
		if hook.Equals(webhook) {
			return
		}
	}
	twl.webhooks = append(twl.webhooks, webhook)
	twl.Save()
}

func (twl *ThresholdWebhookList) UnregisterWebhook(webhook *ThresholdWebhook) {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	out := make([]*ThresholdWebhook, 0, len(twl.webhooks))
	for _, hook := range twl.webhooks {
		if !hook.Equals(webhook) {
			out = append(out, hook)
		}
	}
	twl.webhooks = out
	twl.Save()
}

func (twl *ThresholdWebhookList) List() []*ThresholdWebhook {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	webhooks := make([]*ThresholdWebhook, len(twl.webhooks))
	copy(webhooks, twl.webhooks)
	return webhooks
}

func (twl *ThresholdWebhookList) Evaluate(val float64, data interface{}) {
	webhooks := twl.List()
	for _, webhook := range webhooks {
		hook := webhook
		go hook.Evaluate(val, data)
	}
}

func (twl *ThresholdWebhookList) Monitor(sensor Sensor, interval time.Duration) {
	stop := twl.restart()
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			val, data, err := sensor.Check()
			if err == nil {
				twl.Evaluate(val, data)
			}
		case <-stop:
			ticker.Stop()
			return
		}
	}
}

func (twl *ThresholdWebhookList) realStop() {
	stop := twl.stop
	twl.stop = nil
	if stop != nil {
		stop <- true
		close(stop)
	}
}

func (twl *ThresholdWebhookList) restart() chan bool {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	twl.realStop()
	stop := make(chan bool)
	twl.stop = stop
	return stop
}

func (twl *ThresholdWebhookList) Stop() {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	twl.realStop()
}

