package htp

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"
)

type SyncClient struct {
	client  *http.Client
	context context.Context
	request *http.Request
	tSend   int64
	tRecv   int64
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
				s.tSend = time.Now().UnixNano()
			},
			GotFirstResponseByte: func() {
				s.tRecv = time.Now().UnixNano()
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
	resp.Body.Close()

	dateStr := resp.Header.Get("Date")
	date, err := time.Parse(time.RFC1123, dateStr)
	if err != nil {
		return nil, err
	}

	return &SyncRound{
		Send:    s.tSend,
		Remote:  date.UnixNano(),
		Receive: s.tRecv,
	}, nil
}
