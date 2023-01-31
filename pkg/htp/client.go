package htp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

type SyncClient struct {
	client  *http.Client
	context context.Context
	request *http.Request
	tSend   NanoSec
	tRecv   NanoSec
}

type SyncTrace struct {
	Before func(model *SyncModel) bool
	After  func(model *SyncModel, round *SyncRound) bool
}

func NewSyncClient(host string, timeout time.Duration) (*SyncClient, error) {
	s := &SyncClient{}

	s.client = &http.Client{}
	s.client.Timeout = timeout
	s.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	var err error
	s.request, err = http.NewRequest("HEAD", host, nil)
	if err != nil {
		return nil, err
	}

	s.context = httptrace.WithClientTrace(s.request.Context(),
		&httptrace.ClientTrace{
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				s.tSend = NanoSec(time.Now().UnixNano())
			},
			GotFirstResponseByte: func() {
				s.tRecv = NanoSec(time.Now().UnixNano())
			},
		},
	)

	return s, nil
}

func (s *SyncClient) Round() (*SyncRound, error) {
	resp, err := s.client.Do(s.request.WithContext(s.context))
	if err != nil {
		return nil, err
	}

	// Read and close the body to make sure that the connection can be reused.
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	dateStr := resp.Header.Get("Date")
	date, err := time.Parse(time.RFC1123, dateStr)
	if err != nil {
		return nil, err
	}

	return &SyncRound{
		Send:    s.tSend,
		Remote:  NanoSec(date.UnixNano()),
		Receive: s.tRecv,
	}, nil
}

func (s *SyncClient) Sync(model *SyncModel, trace *SyncTrace) error {
	for {
		if !trace.Before(model) {
			break
		}

		// We wait until the next (estimated) best time to send a request which
		// will reduce the error margin.
		//
		// I.e., we time our request based on the current model so that the
		// server will reply in a "full" second (see README).
		model.Sleep()

		round, err := s.Round()
		if err != nil {
			return err
		}

		if err := model.Update(round); err != nil {
			return err
		}

		if !trace.After(model, round) {
			break
		}
	}

	return nil
}
