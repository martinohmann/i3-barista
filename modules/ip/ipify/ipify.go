package ipify

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/martinohmann/i3-barista/modules/ip"
)

func New() *ip.Module {
	return ip.New(Provider)
}

var Provider = ip.ProviderFunc(func() (net.IP, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		// Request errors most likely indicate that we are offline. Ignore them.
		return nil, nil
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return net.ParseIP(string(buf)), nil
})
