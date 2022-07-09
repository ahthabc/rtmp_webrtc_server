package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/turnserver"
	// "github.com/xiangxud/rtmp_webrtc_server/livekitserver"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	"github.com/xiangxud/rtmp_webrtc_server/util"

	// "github.com/xiangxud/rtmp_webrtc_server/livekitserver"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"
	"github.com/xiangxud/rtmp_webrtc_server/mqtt"
)

func parseConf() {
	if err := config.Parse(); err != nil {
		log.Debugf("cannot parse configuration file:%v", err)
		os.Exit(1)
	}
}
func main() {

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go util.WaitTerm(cancel)
	rand.Seed(time.Now().UTC().UnixNano())
	parseConf()
	m := media_interface.CreateGlobalStreamM(ctx)
	// go livekitserver.Livekit_server(ctx)
	go startRTMPServer(ctx, m)
	go mqtt.StartMqtt(ctx)
	go turnserver.TurnServer() //livekit always turnserver in livekit package
	http.Handle("/", http.FileServer(http.Dir("./web_client")))
	go panic(http.ListenAndServe(":8088", nil))
	<-ctx.Done()
	log.Debugf("stop signal caught, stopping... pid=%d\n", os.Getpid())
}
