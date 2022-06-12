package config

import (
	"context"
	"net/http"

	. "github.com/logrusorgru/aurora"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	"github.com/xiangxud/rtmp_webrtc_server/util"
	"golang.org/x/sync/errgroup"
)

var _ HTTPConfig = (*HTTP)(nil)

type HTTP struct {
	ListenAddr    string
	ListenAddrTLS string
	CertFile      string
	KeyFile       string
	CORS          bool //是否自动添加CORS头
	UserName      string
	Password      string
	mux           *http.ServeMux
}
type HTTPConfig interface {
	InitMux()
	GetHTTPConfig() *HTTP
	Listen(ctx context.Context) error
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}

func (config *HTTP) InitMux() {
	hasOwnTLS := config.ListenAddrTLS != "" && config.ListenAddrTLS != Config.ListenAddrTLS
	hasOwnHTTP := config.ListenAddr != "" && config.ListenAddr != Config.ListenAddr
	if hasOwnTLS || hasOwnHTTP {
		config.mux = http.NewServeMux()
	}
}

func (config *HTTP) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) {
	if config.mux != nil {
		if config.CORS {
			f = util.CORS(f)
		}
		if config.UserName != "" && config.Password != "" {
			f = util.BasicAuth(config.UserName, config.Password, f)
		}
		config.mux.HandleFunc(path, f)
	}
}

func (config *HTTP) GetHTTPConfig() *HTTP {
	return config
}

// ListenAddrs Listen http and https
func (config *HTTP) Listen(ctx context.Context) error {
	var g errgroup.Group
	if config.ListenAddrTLS != "" && (config == &Config.HTTP || config.ListenAddrTLS != Config.ListenAddrTLS) {
		g.Go(func() error {
			log.Info("🌐 https listen at ", Blink(config.ListenAddrTLS))
			return http.ListenAndServeTLS(config.ListenAddrTLS, config.CertFile, config.KeyFile, config.mux)
		})
	}
	if config.ListenAddr != "" && (config == &Config.HTTP || config.ListenAddr != Config.ListenAddr) {
		g.Go(func() error {
			log.Info("🌐 http listen at ", Blink(config.ListenAddr))
			// http.Handle("/", http.FileServer(http.Dir(".")))
			return http.ListenAndServe(config.ListenAddr, config.mux)
		})
	}
	g.Go(func() error {
		<-ctx.Done()
		return ctx.Err()
	})
	return g.Wait()
}
