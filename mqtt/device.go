package mqtt

import (
	"fmt"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
	"github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/livekitclient"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"

	// "github.com/xiangxud/rtmp_webrtc_server/media"
	// media_interface "github.com/xiangxud/rtmp_webrtc_server/media"

	// "github.com/xiangxud/rtmp_webrtc_server/pion_stream"

	enc "github.com/xiangxud/rtmp_webrtc_server/signal"
)

var (
	SDPCh chan *Message
)

const (
	USE_RTP_PUBLISH = true
)

func init() {
	SDPCh = make(chan *Message, 1)
}

type Pull_Stream_From_Device struct {
	Dviceid    string `json:"pullfrom"` //客户端要求服务器拉流的设备标识 也是mqtt topic的唯一值向标识,允许批量
	Devicename string `json:"devicename"`
	// ICEServers
	RoomToken []livekitclient.Token `json:"room_token"` //客户端要求设备加入的房间信息，无此信息则沿用服务器默认的以自身sn创建的房间
}

//客户端发起拉取某设备至多个房间
func DevicePullStream(msg Message) error {
	if msg.Pull_Stream_From_Device == nil {
		return fmt.Errorf("no pull device info")
	}
	// bPullFromDevice := false
	//获取全局的媒体管理对象，里面有房间管理，判断要求的拉流是不是已经存在
	m := media_interface.GetGlobalStreamM()
	deviceroom := m.GetDeviceRoom()
	// deviceroom := livekitclient.GetGlobalDeviceRoom()
	for _, pulldevice := range msg.Pull_Stream_From_Device {
		// bPullFromDevice = true
		deviceid := pulldevice.Dviceid
		devicename := pulldevice.Devicename
		RoomTokenList := pulldevice.RoomToken
		device, _ := deviceroom.GetDevice(deviceid)
		if device == nil {
			device = livekitclient.NewDevice(deviceid, devicename, msg.Describestreamname)
			deviceroom.AddDevice(device)
			log.Debug("Add device to deviceroom->", devicename)
		}
		//本设备未指定房间，可以采用默认房间
		if RoomTokenList == nil {
			// continue
			room := m.GetDefaultRoom()
			device.PRoomList[room.Identity] = room
			continue

		} else {
			//设备对应多个房间
			for _, roomtoken := range RoomTokenList {
				//索引是否已经存在此房间信息
				room := m.GetRoomByName(roomtoken.Identity)
				if room == nil {
					room = livekitclient.NewRoom(mqtt_ctx, &roomtoken)
					room.CreateliveKitRoom(roomtoken.Identity)
					device.PRoomList[room.Identity] = room
					// room.ConnectRoom()
				}

			}
		}
		//拉取请求
		// reqdevpull := Reqdevpull{
		// 	Deviceid:   deviceid,
		// 	ICEServers: msg.ICEServers,
		// }
		msg.SeqID = config.Config.Mqtt.CLIENTID
		go StartPullDeviceStream(deviceid, &msg)
		// GetDevicePullStream(deviceid, &msg)
	}
	//发起拉取设备信令
	// if bPullFromDevice {
	// 	d.GetDevicePullStream()
	// }
	return nil
}
func StartPullDeviceStream(deviceid string, msg *Message) {
	m := media_interface.GetGlobalStreamM()
	deviceroom := m.GetDeviceRoom()
	//查询设备
	device, _ := deviceroom.GetDevice(deviceid)
	if device == nil {
		device = livekitclient.NewDevice(deviceid, deviceid, msg.Describestreamname)
		room := m.GetDefaultRoom()
		device.PRoomList[room.Identity] = room
		deviceroom.AddDevice(device)
	} else {
		device.SetStreamname(msg.Describestreamname)
	}
	NewStream("192.168.0.18", device.Deviceid, device.Streamname, "./", device, msg)
}

type Reqdevpull struct {
	Deviceid   string             `json:"deviceid"`
	ICEServers []webrtc.ICEServer `json:"iceserver"`
}

// func GetDevicePullStream(reqdevpull *Reqdevpull) {
func GetDevicePullStream(deviceid string, msg *Message) {
	req := &Session{}
	req.Type = CMDMSG_SERVER_PULLSTREAMFROM_DEVICE
	req.DeviceId = deviceid
	req.Msg = "pull stream for me"
	msg.Topicprefix = config.Config.Mqtt.SUBTOPIC[:len(config.Config.Mqtt.SUBTOPIC)-2]
	req.Data = enc.Encode(*msg)

	reqmsg := PublishMsg{
		WEB_SEQID: config.Config.Mqtt.CLIENTID,
		Topic:     TOPIC_REQPULL,
		Msg:       req,
		BTodevice: true,
	}
	// {
	// 	t := c.Subscribe(config.Config.Mqtt.SUBTOPIC, config.Config.Mqtt.QOS, h.handle)
	// 	// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
	// 	// in other handlers does cause problems its best to just assume we should not block
	// 	go func() {
	// 		_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
	// 		if t.Error() != nil {
	// 			fmt.Printf("ERROR SUBSCRIBING: %s\n", t.Error())
	// 		} else {
	// 			log.Debug("subscribed to: ", config.Config.Mqtt.SUBTOPIC)
	// 		}
	// 	}()
	// }
	log.Debug("req pull ", msg.SeqID, reqmsg)
	SendMsg(reqmsg)

}
func publishStream(track *webrtc.TrackRemote, streamid, streamname string) (*media_interface.Stream, error) {

	m := media_interface.GetGlobalStreamM()
	pstream, err := m.GetStream(streamname)
	if err != nil {
		log.Debug(err, "error Get current Stream ")
		//return fmt.Errorf("can't initialize codec with %s", err.Error())
	}
	if pstream == nil {
		s := media_interface.Stream{}
		s.InitStream(streamid, streamname, "", "", "RTP") //"RAW") //"RTP")
		// s.InitAudio()
		// s.InitVideo()
		// h.streamname = streamname
		log.Debug("current streamname:", streamname)

		s.SetRemoteTrack(track)
		err := m.AddStream(&s)
		if err != nil {
			log.Debug("addstream error", err)
			return nil, err
		}
		return &s, nil
	} else {
		pstream.SetRemoteTrack(track)
		err := m.AddStream(pstream)
		if err != nil {
			log.Debug("addstream error", err)
			return nil, err
		}
		return pstream, nil
	}

}

//设备发起远程webrtc sdp 协商
func DevicePublishStream(msg Message) {
	m := media_interface.GetGlobalStreamM()
	deviceroom := m.GetDeviceRoom()
	waitaudiochan := make(chan bool, 1)
	waitvideochan := make(chan bool, 1)
	waitstreamtaskchan := make(chan bool, 2)

	// Enable Extension Headers needed for Simulcast
	// media := &webrtc.MediaEngine{}
	// if err := media_interface.RegisterDefaultCodecs(); err != nil {
	// 	panic(err)
	// }
	// for _, extension := range []string{
	// 	"urn:ietf:params:rtp-hdrext:sdes:mid",
	// 	"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
	// 	"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	// } {
	// 	if err := media_interface.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
	// 		panic(err)
	// 	}
	// }
	// // Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// // This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// // this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// // for each PeerConnection.
	// i := &interceptor.Registry{}

	// // Use the default set of Interceptors
	// if err := webrtc.RegisterDefaultInterceptors(media, i); err != nil {
	// 	panic(err)
	// }

	// // Create a new RTCPeerConnection
	// peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(media), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(webrtc.Configuration{
	// 	ICEServers:   msg.ICEServers,
	// 	SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// Create a MediaEngine object to configure the supported codec
	mediaeg := &webrtc.MediaEngine{}
	if err := mediaeg.RegisterDefaultCodecs(); err != nil {
		panic(err)
	}
	// Setup the codecs you want to use.
	// We'll use a VP8 and Opus but you can also define your own
	if err := mediaeg.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/H264", ClockRate: 90000, Channels: 1, SDPFmtpLine: "", RTCPFeedback: nil},
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := mediaeg.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2, SDPFmtpLine: "", RTCPFeedback: nil},
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(mediaeg, i); err != nil {
		panic(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaeg), webrtc.WithInterceptorRegistry(i))

	// Prepare the configuration
	configrtc := webrtc.Configuration{
		ICEServers: msg.ICEServers,
	}

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(configrtc)
	if err != nil {
		panic(err)
	}

	// // Allow us to receive 1 audio track, and 1 video track
	// if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
	// 	panic(err)
	// } else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
	// 	panic(err)
	// }

	// peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
	// 	ICEServers:   msg.ICEServers,
	// 	SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	//设置方向
	peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	//查询设备
	device, _ := deviceroom.GetDevice(msg.SeqID)
	if device == nil {
		device = livekitclient.NewDevice(msg.SeqID, msg.SeqID, msg.Describestreamname)
		room := m.GetDefaultRoom()
		device.PRoomList[room.Identity] = room
		deviceroom.AddDevice(device)
	} else {
		device.SetStreamname(msg.Describestreamname)
	}
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Debug("OnDataChannel", dc.Label())
		startdctask(dc)
	})

	var pstream *media_interface.Stream
	bPublish := false
	// Set a handler for when a new remote track starts, this handler saves buffers to disk as
	// an ivf file, since we could have multiple video tracks we provide a counter.
	// In your application this is where you would handle/process video
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// go func() {
		// 	ticker := time.NewTicker(time.Second * 3)
		// 	for range ticker.C {
		// 		errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
		// 		if errSend != nil {
		// 			fmt.Println(errSend)
		// 		}
		// 	}
		// }()
		// pstream, err := m.GetStream(device.Streamname)
		// if err != nil {
		// 	log.Debug(err, "error Get current Stream ")
		// 	//return fmt.Errorf("can't initialize codec with %s", err.Error())
		// }
		//发布到房间
		if !bPublish && strings.Contains(track.Codec().MimeType, "video") {
			pstream, err = publishStream(track, device.Deviceid, device.Streamname)
			if err != nil {
				log.Debug("publishStream fail", err)
				return
			}
			bPublish = true
		}
		// m := media_interface.GetGlobalStreamM()
		// tiker := time.NewTicker(1 * time.Millisecond)

		codec := track.Codec()
		log.Debug("onTrack", codec)
		if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
			// fmt.Println("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
			// saveToDisk(oggFile, track)

			//转发音频数据包
			go func(audioendchan chan bool, stream *media_interface.Stream) {
				// b := make([]byte, 1500)
				for {
					select {
					case <-audioendchan:
						return
						// Read
					default:
						// if USE_RTP_PUBLISH {
						// 	n, _, readErr := track.Read(b)
						// 	if readErr != nil {
						// 		log.Debug(readErr)
						// 		waitstreamtaskchan <- true
						// 		return
						// 		// panic(readErr)
						// 	}
						// 	for _, proom := range device.PRoomList {
						// 		if proom != nil {
						// 			wrn, err := proom.TrackSendRtpPackets(device.Streamname, "audio", b[:n])
						// 			if wrn != n || err != nil {
						// 				log.Debug(err, "writed ", n)
						// 			}
						// 		}
						// 	}
						// } else {
						pkt, _, err := track.ReadRTP()
						// fmt.Println(filename, pkt.Timestamp)
						if err != nil {
							log.Debug(err)
							return
						}
						if device == nil {
							waitstreamtaskchan <- true
							return
						}
						stream.SendStreamAudioFromWebrtc(pkt.Payload)
						// 	for _, proom := range device.PRoomList {
						// 		if proom != nil {
						// 			proom.TrackSendData(device.Streamname, "audio", pkt.Payload, 20*time.Millisecond)

						// 		}
						// 	}
						// }
					}
				}
			}(waitaudiochan, pstream)
		} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeH264) {
			//转发视频频数据包
			// // Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
			// go func() {
			// 	ticker := time.NewTicker(time.Second * 2)
			// 	for range ticker.C {
			// 		if rtcpErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}}); rtcpErr != nil {
			// 			fmt.Println(rtcpErr)
			// 		}
			// 	}
			// }()
			go func(videoendchan chan bool, stream *media_interface.Stream) {
				// b := make([]byte, 1500)

				for {
					select {
					case <-videoendchan:
						return
					default:
						// case <-tiker.C:
						// Read
						// if USE_RTP_PUBLISH {
						// 	n, _, readErr := track.Read(b)
						// 	if readErr != nil {
						// 		log.Debug(readErr)
						// 		waitstreamtaskchan <- true
						// 		return
						// 		//panic(readErr)
						// 	}
						// 	stream.SendStreamAudio()
						// 	// for _, proom := range device.PRoomList {
						// 	// 	if proom != nil {
						// 	// 		wrn, err := proom.TrackSendRtpPackets(device.Streamname, "video", b[:n])
						// 	// 		if wrn != n || err != nil {
						// 	// 			log.Debug(err, "writed ", n)
						// 	// 		}

						// 	// 	}
						// 	// }
						// } else {
						// time.Sleep(33 * time.Millisecond)
						pkt, _, err := track.ReadRTP()
						if err != nil {
							log.Debug(err)
							return
						}
						h264packet := H264Packet{}
						datas, err := h264packet.GetRTPRawH264(pkt)
						if err != nil {
							//log.Debug(err)
							continue
						}
						if device == nil {
							waitstreamtaskchan <- true
							return
						}
						stream.SendStreamVideo(datas)
						// for _, proom := range device.PRoomList {
						// 	if proom != nil {

						// 		//proom.TrackPublished(device.Streamname)

						// 		proom.TrackSendData(device.Streamname, "video", datas, time.Second/30)
						// 	}
						// }
						// }

					}
				}
			}(waitvideochan, pstream)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			// atomic.AddInt64(&peerConnectionCount, -1)
			if err := peerConnection.Close(); err != nil {
				log.Debug("peerConnection.Close error %v", err)
				return
			}
			log.Debug("peerConnection.Closed")
			for _, proom := range device.PRoomList {
				if proom != nil {
					// proom.Localtracks
					proom.TrackClose(device.Streamname)
				}
			}
			waitstreamtaskchan <- true
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			log.Debug("peerConnection.Connected ")
			// atomic.AddInt64(&peerConnectionCount, 1)
		}
	})
	// Allow us to receive 1 audio track, and 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}

	offer := msg.RtcSession

	device.SetOffer(offer)
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}
	device.SetAnswer(answer)
	<-gatherComplete
	req := &Session{}
	req.Type = CMDMSG_ANSWER
	req.DeviceId = msg.SeqID //config.Config.Mqtt.CLIENTID
	msg.Mode = MODE_ANSWER
	msg.RtcSession = *peerConnection.LocalDescription()
	req.Data = enc.Encode(msg)
	answermsg := PublishMsg{
		WEB_SEQID: config.Config.Mqtt.CLIENTID, //msg.SeqID,
		Topic:     TOPIC_ANSWER,
		Msg:       req,
		BTodevice: true,
	}
	log.Debug("answer ", msg)
	SendMsg(answermsg)

	// if USE_RTP_PUBLISH {
	// 	for _, proom := range device.PRoomList {
	// 		if proom != nil {
	// 			// proom.Localtracks
	// 			proom.RTPTrackPublished(device.Streamname)
	// 		}
	// 	}
	// } else {
	// 	for _, proom := range device.PRoomList {
	// 		if proom != nil {
	// 			// proom.Localtracks
	// 			proom.TrackPublished(device.Streamname)
	// 		}
	// 	}
	// }
	streamexit := <-waitstreamtaskchan
	waitaudiochan <- true
	waitvideochan <- true
	log.Debug("stream exit ", streamexit)
	// select {}
	// return nil
}

func startdctask(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		err := dc.SendText("please input command")
		if err != nil {
			log.Debug("write data error:", err)
			dc.Close()
		}
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Debug("dc msg:", msg.Data)
		dc.SendText("ok")
	})
	dc.OnClose(func() {
		//dc.ID()
		log.Debug("Close Control socket")
	})
}

// func DevicePublishStream1(msg Message) {
// 	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.
// 	m := media_interface.GetGlobalStreamM()
// 	deviceroom := m.GetDeviceRoom()
// 	fmt.Println(deviceroom)
// 	// Prepare the configuration
// 	config_webrtc := webrtc.Configuration{
// 		ICEServers: msg.ICEServers,
// 		// ICEServers: []webrtc.ICEServer{
// 		// 	{
// 		// 		URLs: []string{"stun:stun.l.google.com:19302"},
// 		// 	},
// 	}

// 	// Enable Extension Headers needed for Simulcast
// 	media := &webrtc.MediaEngine{}
// 	if err := media_interface.RegisterDefaultCodecs(); err != nil {
// 		panic(err)
// 	}
// 	for _, extension := range []string{
// 		"urn:ietf:params:rtp-hdrext:sdes:mid",
// 		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
// 		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
// 	} {
// 		if err := media_interface.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
// 			panic(err)
// 		}
// 	}

// 	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
// 	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
// 	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
// 	// for each PeerConnection.
// 	i := &interceptor.Registry{}

// 	// Use the default set of Interceptors
// 	if err := webrtc.RegisterDefaultInterceptors(media, i); err != nil {
// 		panic(err)
// 	}

// 	// Create a new RTCPeerConnection
// 	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(media), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(config_webrtc)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer func() {
// 		if cErr := peerConnection.Close(); cErr != nil {
// 			fmt.Printf("cannot close peerConnection: %v\n", cErr)
// 		}
// 	}()

// 	outputTracks := map[string]*webrtc.TrackLocalStaticRTP{}

// 	// Create Track that we send video back to browser on
// 	outputTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video_q", "pion_q")
// 	if err != nil {
// 		panic(err)
// 	}
// 	outputTracks["q"] = outputTrack

// 	outputTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video_h", "pion_h")
// 	if err != nil {
// 		panic(err)
// 	}
// 	outputTracks["h"] = outputTrack

// 	outputTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video_f", "pion_f")
// 	if err != nil {
// 		panic(err)
// 	}
// 	outputTracks["f"] = outputTrack

// 	// Add this newly created track to the PeerConnection
// 	if _, err = peerConnection.AddTrack(outputTracks["q"]); err != nil {
// 		panic(err)
// 	}
// 	if _, err = peerConnection.AddTrack(outputTracks["h"]); err != nil {
// 		panic(err)
// 	}
// 	if _, err = peerConnection.AddTrack(outputTracks["f"]); err != nil {
// 		panic(err)
// 	}

// 	// Read incoming RTCP packets
// 	// Before these packets are returned they are processed by interceptors. For things
// 	// like NACK this needs to be called.
// 	processRTCP := func(rtpSender *webrtc.RTPSender) {
// 		rtcpBuf := make([]byte, 1500)
// 		for {
// 			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
// 				return
// 			}
// 		}
// 	}
// 	for _, rtpSender := range peerConnection.GetSenders() {
// 		go processRTCP(rtpSender)
// 	}

// 	// Wait for the offer to be pasted
// 	offer := msg.RtcSession //webrtc.SessionDescription{}
// 	// signal.Decode(signal.MustReadStdin(), &offer)

// 	if err = peerConnection.SetRemoteDescription(offer); err != nil {
// 		panic(err)
// 	}
// 	// {
// 	// 		//查询设备
// 	// 		device, _ := deviceroom.GetDevice(msg.SeqID)
// 	// 		if device == nil {
// 	// 			device = livekitclient.NewDevice(msg.SeqID, msg.SeqID, msg.Describestreamname)
// 	// 			room := m.GetDefaultRoom()
// 	// 			device.PRoomList[room.Identity] = room
// 	// 			deviceroom.AddDevice(device)
// 	// 		} else {
// 	// 			device.SetStreamname(msg.Describestreamname)
// 	// 		}
// 	// 		// Set a handler for when a new remote track starts, this handler saves buffers to disk as
// 	// 		// an ivf file, since we could have multiple video tracks we provide a counter.
// 	// 		// In your application this is where you would handle/process video
// 	// 		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 	// 			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
// 	// 			// go func() {
// 	// 			// 	ticker := time.NewTicker(time.Second * 3)
// 	// 			// 	for range ticker.C {
// 	// 			// 		errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
// 	// 			// 		if errSend != nil {
// 	// 			// 			fmt.Println(errSend)
// 	// 			// 		}
// 	// 			// 	}
// 	// 			// }()

// 	// 			codec := track.Codec()
// 	// 			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
// 	// 				// fmt.Println("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
// 	// 				// saveToDisk(oggFile, track)

// 	// 				//转发音频数据包
// 	// 				go func() {
// 	// 					b := make([]byte, 1500)
// 	// 					for {
// 	// 						// Read
// 	// 						n, _, readErr := track.Read(b)
// 	// 						if readErr != nil {
// 	// 							panic(readErr)
// 	// 						}

// 	// 						for _, proom := range device.PRoomList {
// 	// 							if proom != nil {
// 	// 								// proom.TrackPublished(device.Streamname)
// 	// 								// track.Read
// 	// 								proom.TrackSendData(device.Streamname, "audio", b[:n], 20*time.Millisecond)
// 	// 							}
// 	// 						}
// 	// 					}
// 	// 				}()
// 	// 			} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeH264) {
// 	// 				//转发视频频数据包
// 	// 				go func() {
// 	// 					b := make([]byte, 1500)
// 	// 					for {
// 	// 						// Read
// 	// 						n, _, readErr := track.Read(b)
// 	// 						if readErr != nil {
// 	// 							panic(readErr)
// 	// 						}

// 	// 						for _, proom := range device.PRoomList {
// 	// 							if proom != nil {
// 	// 								// proom.TrackPublished(device.Streamname)
// 	// 								// track.Read
// 	// 								track.PayloadType()
// 	// 								proom.TrackSendData(device.Streamname, "video", b[:n], time.Second/30)
// 	// 							}
// 	// 						}
// 	// 					}
// 	// 				}()
// 	// 			}
// 	// 		})
// 	// }
// 	// Set a handler for when a new remote track starts
// 	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 		fmt.Println("Track has started")

// 		go func() {
// 			ticker := time.NewTicker(3 * time.Second)
// 			for range ticker.C {
// 				fmt.Printf("Sending pli for stream with rid: %q, ssrc: %d\n", track.RID(), track.SSRC())
// 				if writeErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}}); writeErr != nil {
// 					fmt.Println(writeErr)
// 				}
// 			}
// 		}()

// 		// Read incoming RTCP packets
// 		// Before these packets are returned they are processed by interceptors. For things
// 		// like TWCC and RTCP Reports this needs to be called.
// 		go func() {
// 			rtcpBuf := make([]byte, 1500)
// 			for {
// 				if _, _, rtcpErr := receiver.Read(rtcpBuf); rtcpErr != nil {
// 					return
// 				}
// 			}
// 		}()

// 		// Start reading from all the streams and sending them to the related output track
// 		rid := track.RID()
// 		for {
// 			// Read RTP packets being sent to Pion
// 			packet, _, readErr := track.ReadRTP()
// 			if readErr != nil {
// 				panic(readErr)
// 			}

// 			if writeErr := outputTracks[rid].WriteRTP(packet); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
// 				panic(writeErr)
// 			}
// 		}
// 	})

// 	// Set the handler for Peer connection state
// 	// This will notify you when the peer has connected/disconnected
// 	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
// 		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

// 		if s == webrtc.PeerConnectionStateFailed {
// 			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
// 			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
// 			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
// 			fmt.Println("Peer Connection has gone to failed exiting")
// 			os.Exit(0)
// 		}
// 	})

// 	// Create an answer
// 	answer, err := peerConnection.CreateAnswer(nil)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create channel that is blocked until ICE Gathering is complete
// 	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

// 	// Sets the LocalDescription, and starts our UDP listeners
// 	err = peerConnection.SetLocalDescription(answer)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Block until ICE Gathering is complete, disabling trickle ICE
// 	// we do this because we only can exchange one signaling message
// 	// in a production application you should exchange ICE Candidates via OnICECandidate
// 	<-gatherComplete

// 	// Output the answer in base64 so we can paste it in browser
// 	fmt.Println(signal.Encode(*peerConnection.LocalDescription()))
// 	req := &Session{}
// 	req.Type = CMDMSG_ANSWER
// 	req.DeviceId = msg.SeqID //config.Config.Mqtt.CLIENTID

// 	req.Data = enc.Encode(*peerConnection.LocalDescription())
// 	answermsg := PublishMsg{
// 		WEB_SEQID: config.Config.Mqtt.CLIENTID, //msg.SeqID,
// 		Topic:     TOPIC_ANSWER,
// 		Msg:       req,
// 	}
// 	log.Debugf("answer %s", msg.SeqID)
// 	SendMsg(answermsg)
// 	// Block forever
// 	select {}
// }

// func GetDevicePullStream1(deviceid string, msg *Message) { //nolint:gocognit
// 	m := media_interface.GetGlobalStreamM()
// 	deviceroom := m.GetDeviceRoom()
// 	var candidatesMux sync.Mutex
// 	pendingCandidates := make([]*webrtc.ICECandidate, 0)
// 	// We make our own mediaEngine so we can place the sender's codecs in it.  This because we must use the
// 	// dynamic media type from the sender in our answer. This is not required if we are the offerer
// 	mediaEngine := &webrtc.MediaEngine{}
// 	mediaEngine.RegisterDefaultCodecs()

// 	// Create a new RTCPeerConnection
// 	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
// 	// Everything below is the Pion WebRTC API! Thanks for using it ❤️.

// 	// Prepare the configuration
// 	config_webrtc := webrtc.Configuration{
// 		ICEServers: msg.ICEServers,
// 		// ICEServers: []webrtc.ICEServer{
// 		// 	{
// 		// 		URLs: []string{"stun:stun.l.google.com:19302"},
// 		// 	},
// 		// },
// 	}

// 	// Create a new RTCPeerConnection
// 	peerConnection, err := api.NewPeerConnection(config_webrtc)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer func() {
// 		if cErr := peerConnection.Close(); cErr != nil {
// 			fmt.Printf("cannot close peerConnection: %v\n", cErr)
// 		}
// 	}()

// 	// if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
// 	// 	panic(err)
// 	// }
// 	// if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
// 	// 	panic(err)
// 	// }
// 	//设置方向
// 	peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
// 		Direction: webrtc.RTPTransceiverDirectionRecvonly,
// 	})
// 	peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
// 		Direction: webrtc.RTPTransceiverDirectionRecvonly,
// 	})
// 	gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)
// 	// Create an offer to send to the other process
// 	offer, err := peerConnection.CreateOffer(nil)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Sets the LocalDescription, and starts our UDP listeners
// 	// Note: this will start the gathering of ICE candidates
// 	if err = peerConnection.SetLocalDescription(offer); err != nil {
// 		panic(err)
// 	}
// 	// When an ICE candidate is available send to the other Pion instance
// 	// the other Pion instance will add this candidate by calling AddICECandidate
// 	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
// 		if c == nil {
// 			return
// 		}

// 		candidatesMux.Lock()
// 		defer candidatesMux.Unlock()

// 		desc := peerConnection.RemoteDescription()
// 		if desc == nil {
// 			pendingCandidates = append(pendingCandidates, c)
// 		}
// 		// } else if onICECandidateErr := signalCandidate(*answerAddr, c); onICECandidateErr != nil {
// 		// 	panic(onICECandidateErr)
// 		// }
// 	})

// 	{
// 		//查询设备
// 		device, _ := deviceroom.GetDevice(msg.SeqID)
// 		if device == nil {
// 			device = livekitclient.NewDevice(msg.SeqID, msg.SeqID, msg.Describestreamname)
// 			room := m.GetDefaultRoom()
// 			device.PRoomList[room.Identity] = room
// 			deviceroom.AddDevice(device)
// 		} else {
// 			device.SetStreamname(msg.Describestreamname)
// 		}
// 		// Set a handler for when a new remote track starts, this handler saves buffers to disk as
// 		// an ivf file, since we could have multiple video tracks we provide a counter.
// 		// In your application this is where you would handle/process video
// 		peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 			// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
// 			// go func() {
// 			// 	ticker := time.NewTicker(time.Second * 3)
// 			// 	for range ticker.C {
// 			// 		errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
// 			// 		if errSend != nil {
// 			// 			fmt.Println(errSend)
// 			// 		}
// 			// 	}
// 			// }()

// 			codec := track.Codec()
// 			if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
// 				// fmt.Println("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
// 				// saveToDisk(oggFile, track)

// 				//转发音频数据包
// 				go func() {
// 					b := make([]byte, 1500)
// 					for {
// 						// Read
// 						n, _, readErr := track.Read(b)
// 						if readErr != nil {
// 							panic(readErr)
// 						}

// 						for _, proom := range device.PRoomList {
// 							if proom != nil {
// 								// proom.TrackPublished(device.Streamname)
// 								// track.Read
// 								proom.TrackSendData(device.Streamname, "audio", b[:n], 20*time.Millisecond)
// 							}
// 						}
// 					}
// 				}()
// 			} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeH264) {
// 				//转发视频频数据包
// 				go func() {
// 					b := make([]byte, 1500)
// 					for {
// 						// Read
// 						n, _, readErr := track.Read(b)
// 						if readErr != nil {
// 							panic(readErr)
// 						}

// 						for _, proom := range device.PRoomList {
// 							if proom != nil {
// 								// proom.TrackPublished(device.Streamname)
// 								// track.Read
// 								track.PayloadType()
// 								proom.TrackSendData(device.Streamname, "video", b[:n], time.Second/30)
// 							}
// 						}
// 					}
// 				}()
// 			}
// 		})
// 	}
// 	go func() {
// 		answermsg := <-SDPCh
// 		answersdp := (*answermsg).RtcSession
// 		if sdpErr := peerConnection.SetRemoteDescription(answersdp); sdpErr != nil {
// 			panic(sdpErr)
// 		}
// 	}()
// 	// // A HTTP handler that processes a SessionDescription given to us from the other Pion process
// 	// http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
// 	// 	sdp := webrtc.SessionDescription{}
// 	// 	if sdpErr := json.NewDecoder(r.Body).Decode(&sdp); sdpErr != nil {
// 	// 		panic(sdpErr)
// 	// 	}

// 	// 	if sdpErr := peerConnection.SetRemoteDescription(sdp); sdpErr != nil {
// 	// 		panic(sdpErr)
// 	// 	}

// 	// 	candidatesMux.Lock()
// 	// 	defer candidatesMux.Unlock()

// 	// 	for _, c := range pendingCandidates {
// 	// 		if onICECandidateErr := signalCandidate(*answerAddr, c); onICECandidateErr != nil {
// 	// 			panic(onICECandidateErr)
// 	// 		}
// 	// 	}
// 	// })
// 	// Start HTTP server that accepts requests from the answer process
// 	// go func() { panic(http.ListenAndServe(*offerAddr, nil)) }()

// 	// Create a datachannel with label 'data'
// 	// dataChannel, err := peerConnection.CreateDataChannel("data", nil)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// Set the handler for Peer connection state
// 	// This will notify you when the peer has connected/disconnected
// 	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
// 		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

// 		if s == webrtc.PeerConnectionStateFailed {
// 			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
// 			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
// 			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
// 			fmt.Println("Peer Connection has gone to failed exiting")
// 			os.Exit(0)
// 		}
// 	})

// 	// Register channel opening handling
// 	// dataChannel.OnOpen(func() {
// 	// 	fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", dataChannel.Label(), dataChannel.ID())

// 	// 	for range time.NewTicker(5 * time.Second).C {
// 	// 		message := signal.RandSeq(15)
// 	// 		fmt.Printf("Sending '%s'\n", message)

// 	// 		// Send the message as text
// 	// 		sendTextErr := dataChannel.SendText(message)
// 	// 		if sendTextErr != nil {
// 	// 			panic(sendTextErr)
// 	// 		}
// 	// 	}
// 	// })

// 	// // Register text message handling
// 	// dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
// 	// 	fmt.Printf("Message from DataChannel '%s': '%s'\n", dataChannel.Label(), string(msg.Data))
// 	// })

// 	req := &Session{}
// 	req.Type = CMDMSG_OFFER
// 	req.DeviceId = deviceid //msg.SeqID //config.Config.Mqtt.CLIENTID
// 	msg.SeqID = config.Config.Mqtt.CLIENTID
// 	msg.RtcSession = offer
// 	msg.Mode = MODE_OFFER
// 	msg.Video = true //
// 	msg.Audio = true //
// 	msg.Describestreamname = "testdevicestream"
// 	req.Data = enc.Encode(msg)
// 	answermsg := PublishMsg{
// 		WEB_SEQID: config.Config.Mqtt.CLIENTID, //msg.SeqID,
// 		Topic:     TOPIC_OFFER,
// 		Msg:       req,
// 		BTodevice: true,
// 	}
// 	log.Debugf("answer %s", msg.SeqID)
// 	SendMsg(answermsg)
// 	<-gatherCompletePromise
// 	// Send our offer to the HTTP server listening in the other process
// 	// payload, err := json.Marshal(offer)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// resp, err := http.Post(fmt.Sprintf("http://%s/sdp", *answerAddr), "application/json; charset=utf-8", bytes.NewReader(payload)) // nolint:noctx
// 	// if err != nil {
// 	// 	panic(err)
// 	// } else if err := resp.Body.Close(); err != nil {
// 	// 	panic(err)
// 	// }

// 	// Block forever
// 	select {}
// }
