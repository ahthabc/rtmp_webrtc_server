package main

import (
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/xiangxud/rtmp_webrtc_server/config"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"
	"github.com/xiangxud/rtmp_webrtc_server/mqtt"
	turn "github.com/xiangxud/rtmp_webrtc_server/turnserver"
)

func parseConf() {
	if err := config.Parse(); err != nil {
		logger.Debugf("cannot parse configuration file:%v", err)
		os.Exit(1)
	}
}
func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	rand.Seed(time.Now().UTC().UnixNano())
	parseConf()
	m := media_interface.CreateGlobalStreamM()

	go startRTMPServer(m)
	go mqtt.StartMqtt()
	go turn.TurnServer()
	http.Handle("/", http.FileServer(http.Dir("./web_client")))
	go panic(http.ListenAndServe(":8080", nil))
	<-c
	logger.Debugf("stop signal caught, stopping... pid=%d\n", os.Getpid())
}
