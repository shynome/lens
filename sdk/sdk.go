package sdk

import (
	"bytes"
	"io"
	"net/http"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type Sdk struct {
	Endpoint string
	client   *http.Client
}

func New(endpoint string) *Sdk {
	return &Sdk{
		Endpoint: endpoint,
		client:   http.DefaultClient,
	}
}
func (s *Sdk) Call(input []byte) (output []byte, err error) {
	defer err2.Return(&err)

	req := try.To1(
		http.NewRequest(http.MethodPost, s.Endpoint, bytes.NewReader(input)))

	resp := try.To1(
		s.client.Do(req))

	output = try.To1(
		io.ReadAll(resp.Body))

	return
}

func (s *Sdk) Dial(id string, result []byte) (err error) {
	defer err2.Return(&err)
	req := try.To1(
		http.NewRequest(http.MethodDelete, s.Endpoint, bytes.NewReader(result)))
	req.Header.Set("X-Event-Id", id)
	_ = try.To1(
		s.client.Do(req))
	return
}
