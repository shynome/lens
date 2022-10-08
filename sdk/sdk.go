package sdk

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type Sdk struct {
	Endpoint string
	Client   *http.Client
}

func New(endpoint string) *Sdk {
	return &Sdk{
		Endpoint: endpoint,
		Client:   http.DefaultClient,
	}
}
func (s *Sdk) Call(topic string, input []byte) (output []byte, err error) {
	defer err2.Return(&err)

	endpoint := try.To1(WithTopic(s.Endpoint, topic))
	req := try.To1(http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(input)))
	resp := try.To1(s.Client.Do(req))
	try.To(CheckResp(resp))
	output = try.To1(io.ReadAll(resp.Body))

	return
}

func (s *Sdk) Dial(topic string, id string, result []byte) (err error) {
	defer err2.Return(&err)

	endpoint := try.To1(WithTopic(s.Endpoint, topic))
	req := try.To1(http.NewRequest(http.MethodDelete, endpoint, bytes.NewReader(result)))
	req.Header.Set("X-Event-Id", id)
	resp := try.To1(s.Client.Do(req))
	try.To(CheckResp(resp))

	return
}

func WithTopic(endpoint string, topic string) (rEndpoint string, err error) {
	defer err2.Return(&err)

	if topic == "" {
		return endpoint, nil
	}

	u := try.To1(url.Parse(endpoint))
	q := u.Query()
	q.Set("t", topic)
	u.RawQuery = q.Encode()

	rEndpoint = u.String()

	return
}

func CheckResp(resp *http.Response) (err error) {
	defer err2.Return(&err)
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		errText := try.To1(io.ReadAll(resp.Body))
		err = fmt.Errorf("server err. code: %v. content: %s", resp.StatusCode, errText)
		return
	}
	return
}
