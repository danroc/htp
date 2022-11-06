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
