//go:build !debug
// +build !debug

package http2_util

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/nycu-ucr/gonet/http"
	"github.com/nycu-ucr/net/http2"
	"github.com/nycu-ucr/net/http2/h2c"
	"github.com/nycu-ucr/net/http2/onvm2c"
	"github.com/pkg/errors"
)

const (
	USE_ONVM_CONN     = true
	USE_ONVM_XIO_CONN = true
	USE_ONVM_HANDLER  = true
)

// NewServer returns a server instance with HTTP/2.0 and HTTP/2.0 cleartext support
// If this function cannot open or create the secret log file,
// **it still returns server instance** but without the secret log and error indication
func NewServer(bindAddr string, preMasterSecretLogPath string, handler http.Handler) (server *http.Server, err error) {
	if handler == nil {
		return nil, errors.New("server needs handler to handle request")
	}

	h2s := &http2.Server{
		// TODO: extends the idle time after re-use openapi client
		// IdleTimeout: 1 * time.Millisecond,
		IdleTimeout: 10 * time.Second,
	}

	if USE_ONVM_CONN {
		// ONVM Connection
		if USE_ONVM_HANDLER {
			// ONVM HTTP handler
			server = &http.Server{
				USING_ONVM_SOCKET:     USE_ONVM_CONN,
				USING_ONVM_XIO_SOCKET: USE_ONVM_XIO_CONN,
				Addr:                  bindAddr,
				Handler:               onvm2c.NewHandler(handler, h2s),
			}
		} else {
			// TCP HTTP handler
			server = &http.Server{
				USING_ONVM_SOCKET:     USE_ONVM_CONN,
				USING_ONVM_XIO_SOCKET: USE_ONVM_XIO_CONN,
				Addr:                  bindAddr,
				Handler:               h2c.NewHandler(handler, h2s),
			}
		}
	} else {
		// TCP Connection
		server = &http.Server{
			USING_ONVM_SOCKET: USE_ONVM_CONN,
			Addr:              bindAddr,
			Handler:           h2c.NewHandler(handler, h2s),
		}
	}

	if preMasterSecretLogPath != "" {
		preMasterSecretFile, err := os.OpenFile(preMasterSecretLogPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return server, fmt.Errorf("create pre-master-secret log [%s] fail: %s", preMasterSecretLogPath, err)
		}
		server.TLSConfig = &tls.Config{
			KeyLogWriter: preMasterSecretFile,
		}
	}

	return
}
