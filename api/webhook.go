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

type WebhookWithID struct {
	*events.Webhook
	id int64
}

type WebhookList struct {
	fn string
	eventSink events.EventSink
	mutex *sync.Mutex
	hooks map[string]map[]*WebhookWithID
}

func NewWebhookList(cfg *Config, eventSink events.EventSink) (*WebhookList, error) {
	fn := cfg.Abs("var/webhooks.json")
	hooks := map[string][]*WebhookWithID{}
	f, err := os.Open(fn)
	if err == nil {
		defer f.Close()
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &hooks)
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	whl := &WebhookList{
		eventSink: eventSink,
		fn: fn,
		mutex: &sync.Mutex{},
		hooks: map[string][]*WebhookWithID{},
	}
	onRemove := events.NewEventHandler(func(ev events.Event) error {
		evData, ok := ev.GetData().(*events.ListenerMeta)
		if !ok {
			return nil
		}
		return whl.remove(evData.EventType, evData.HandlerID)
	})
	eventSink.AddEventListener(events.EventTypeHandlerRemoved, onRemove)
	for eventType, webhooks := range hooks {
		for _, webhook := range webhooks {
			handler := webhook.Handler()
			webhook.id = handler.ID()
			eventSink.AddEventListener(eventType, handler)
		}
	}
	return whl, nil
}

func (whl *WebhookList) Save() error {
	tmpfn := whl.fn+".tmp"
	f, err := os.Create(tmpfn)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	whl.mutex.Lock()
	err = enc.Encode(whl.hooks)
	whl.mutex.Unlock()
	if err != nil {
		f.Close()
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tmpfn, whl.fn)
	return err
}

func (whl *WebhookList) remove(eventType string, handlerId int64) error {
	whl.mutex.Lock()
	hooks := whl.hooks[eventType]
	out := make([]*WebhookWithID, 0, len(hooks))
	for _, hook := range hooks {
		if hook.id != handlerId {
			out = append(out, hook)
		}
	}
	if len(out) == 0 {
		delete(whl.hooks, eventType)
	} else {
		whl.hooks[eventType] = out
	}
	whl.mutex.Unlock()
	return whl.Save()
}

func (whl *WebhookList) add(eventType string, hook *events.Webhook) error {
	handler := hook.Handler()
	whl.eventSink.AddEventListener(eventType, handler)
	whl.mutex.Lock()
	whl.hooks[eventType] = append(whl.hooks[eventType], &WebhookWithID{hook, handler.ID()})
	whl.mutex.Unlock()
	return whl.Save()
}

func (whl *WebhookList) HandleAddWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	eventType := r.URL.Query().Get("event_type")
	if eventType == "" {
		return nil, httpserver.BadRequest.New("missing event type")
	}
	hook := &events.Webhook{}
	err := httpserver.ReadJSON(r, hook)
	if err != nil {
		return nil, err
	}
	err := whl.add(eventType, hook)
	if err != nil {
		return nil, err
	}
	return hook, nil
}

func (whl *WebhookList) HandleRemoveWebhook(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	eventType := r.URL.Query().Get("event_type")
	if eventType == "" {
		return nil, httpserver.BadRequest.New("missing event type")
	}
	hook := &events.Webhook{}
	err := httpserver.ReadJSON(r, hook)
	if err != nil {
		return nil, err
	}
	ids := []int64{}
	whl.mutex.Lock()
	hooks := whl.hooks[eventType]
	whl.mutex.Unlock()
	for _, xhook := range hooks {
		if xhook.Equals(hook) {
			ids = append(ids, xhook.id)
		}
	}
	for _, id := range ids {
		whl.eventSink.RemoveEventListener(eventType, events.HandlerReference(id))
	}
	return hook, nil
}

func (whl *WebhookList) HandleListWebhooks(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	eventType := r.URL.Query().Get("event_type")
	var out []byte
	var err error
	whl.mutex.Lock()
	if eventType == "" {
		out, err = json.Marshal(whl.hooks)
	} else {
		out, err = json.Marshal(whl.hooks[eventType])
	}
	whl.mutex.Unlock()
	return out, err
}

func (whl *WebhookList) HandleListEventTypes(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return whl.eventSink.ListEventTypes(), nil
}

func (whl *WebhookList) HandleEventLog(w http.ResponseWriter, w *http.Request) (interface{}, error) {
	return whl.eventSink.Log(), nil
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

