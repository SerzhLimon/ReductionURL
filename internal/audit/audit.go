package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/SerzhLimon/ReductionURL/internal/config"
	"github.com/sirupsen/logrus"
)

const (
	Shorten string = "shorten"
	Follow  string = "follow"
)

type Event struct {
	Ts     time.Time `json:"ts"`
	Action string    `json:"action"`
	UserID int       `json:"user_id"`
	URL    string    `json:"url"`
}

type Observer interface {
	Run(ctx context.Context)
	Update(event Event)
}

type Audit struct {
	cfg       *config.Config
	eventChan chan Event
	
	hasFile   bool
	hasURL    bool
}

func New(cfg *config.Config) (Observer, error) {
	if cfg == nil || cfg.Opts == nil {
		return nil, fmt.Errorf("NewAudit(): empty config")
	}
	if cfg.Opts.AuditFile == "" && cfg.Opts.AuditURL == "" {
		return nil, fmt.Errorf("NewAudit(): empty config")
	}

	return &Audit{
		cfg:       cfg,
		hasFile:   cfg.Opts.AuditFile != "",
		hasURL:    cfg.Opts.AuditURL != "",
		eventChan: make(chan Event),
	}, nil
}

func (a *Audit) Update(event Event) {
	a.eventChan <- event
}

func (a *Audit) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logrus.Info("audit stopped")
			return
		case event := <-a.eventChan:
			go a.sendEvent(event)
		}
	}
}

func (a *Audit) sendEvent(event Event) {
	if a.hasFile {
		if err := a.saveToFile(event); err != nil {
			logrus.Warnf("sendEvent() %v", err)
		}
	}
	if a.hasURL {
		if err := a.sendToHost(event); err != nil {
			logrus.Warnf("sendEvent() %v", err)
		}
	}
}

func (a *Audit) saveToFile(event Event) error {
	dir := filepath.Dir(a.cfg.Opts.AuditFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dataEvent, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(a.cfg.Opts.AuditFile, dataEvent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (a *Audit) sendToHost(event Event) error {

	dataEvent, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	addr, _ := url.ParseRequestURI(a.cfg.Opts.AuditURL)
	body := bytes.NewReader(dataEvent)
	
	req, err := http.NewRequest(http.MethodPost, addr.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(body.Len())

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	_, err = client.Do(req)

	return err
}

func CreateEvent(userID int, action, URL string) Event {
	return Event{
		Ts: time.Now(),
		Action: action,
		URL: URL,
		UserID: userID,
	}
}