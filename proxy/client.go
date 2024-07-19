package proxy

import (
	"net/url"

	"github.com/rollkit/go-sequencing"
)

func NewClient(uri string) (sequencing.Sequencer, error) {
	addr, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	switch addr.Scheme {
	case "grpc":
	default:
		return nil, fmt.Errorf("unknown url scheme '%s'", addr.Scheme)
}
