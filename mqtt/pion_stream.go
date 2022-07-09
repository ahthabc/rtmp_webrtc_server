package mqtt

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264writer"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/livekitclient"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"

	// "github.com/xiangxud/rtmp_webrtc_server/mqtt"
	enc "github.com/xiangxud/rtmp_webrtc_server/signal"
)

var Streams sync.Map

func Find(host, room string) bool {
	key := "webrtc://" + host + "/" + room
	if _, ok := Streams.Load(key); ok {
		return true
	}
	return false
}

func LoadAndDelStream(host, room string) *Stream {
	key := "webrtc://" + host + "/" + room
	if v, ok := Streams.LoadAndDelete(key); ok {
		if pion_stream, ok := v.(*Stream); ok {
			return pion_stream
		}
		return nil
	}
	return nil
}

type Stream struct {
	Host          string
	Room          string
	Display       string
	rtcUrl        string
	savePath      string
	ctx           context.Context
	cancel        context.CancelFunc
	pc            *webrtc.PeerConnection
	hasAudioTrack bool
	hasVideoTrack bool
	bAudioStop    bool
	bVideoStop    bool
	bDatachanel   bool
	videoFinish   chan struct{}
	audioFinish   chan struct{}
	device        *livekitclient.Device
}

var (
	bWrite = false //转发和存盘切换
)

func (pps *Stream) onTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver, stream *media_interface.Stream) error {
	// Send a PLI on an interval so that the publisher is pushing a keyframe
	codec := track.Codec()

	trackDesc := fmt.Sprintf("channels=%v", codec.Channels)
	if track.Kind() == webrtc.RTPCodecTypeVideo {
		trackDesc = fmt.Sprintf("fmtp=%v", codec.SDPFmtpLine)
	}
	logrus.Infof("Got track %v, pt=%v tbn=%v, %v", codec.MimeType, codec.PayloadType, codec.ClockRate, trackDesc)
	if bWrite {
		var err error
		if codec.MimeType == "audio/opus" {
			var da media.Writer
			defer func() {
				if da != nil {
					da.Close()
				}
			}()
			audiopath := pps.savePath + pps.Display + "_audio.ogg"
			if da, err = oggwriter.New(audiopath, codec.ClockRate, codec.Channels); err != nil {
				return errors.Wrapf(err, "创建"+audiopath+"失败")
			}
			pps.hasAudioTrack = true
			logrus.Infof("Open ogg writer file=%v , tbn=%v, channels=%v", audiopath, codec.ClockRate, codec.Channels)
			if err = pps.writeTrackToDisk(da, track); err != nil {
				return err
			}
			pps.audioFinish <- struct{}{}
		} else if codec.MimeType == "video/H264" {
			var dv_h264 media.Writer
			videopath := pps.savePath + pps.Display + "_video.h264"

			if dv_h264, err = h264writer.New(videopath); err != nil {
				return err
			}
			logrus.Infof("Open h264 writer file=%v", videopath)
			pps.hasVideoTrack = true
			if err = pps.writeTrackToDisk(dv_h264, track); err != nil {
				return err
			}
			pps.audioFinish <- struct{}{}
		} else {
			logrus.Warnf("Ignore track %v pt=%v", codec.MimeType, codec.PayloadType)
		}
	} else {
		// codec := track.Codec()
		if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
			// fmt.Println("Got Opus track, saving to disk as output.opus (48 kHz, 2 channels)")
			// saveToDisk(oggFile, track)

			//转发音频数据包
			//go func() {
			b := make([]byte, 1500)
			pps.bAudioStop = false
			for !pps.bAudioStop {
				if !stream.IsRtpStream() {
					pkt, _, err := track.ReadRTP()
					// fmt.Println(filename, pkt.Timestamp)
					if err != nil {
						log.Debug(err)
						break
					}

					stream.SendStreamAudioFromWebrtc(pkt.Payload)
				} else {

					n, _, readErr := track.Read(b)
					if readErr != nil {
						// log.Debug(readErr)
						break
						//panic(readErr)
					}
					stream.SendStreamAudioFromWebrtc(b[:n])

				}
				// stream.SendStreamAudioFromWebrtc(pkt.Payload)
				// // for _, proom := range pps.device.PRoomList {
				// // 	if proom != nil {
				// // 		// proom.TrackPublished(device.Streamname)
				// // 		// track.Read
				// // 		// if pps.device.Streamname!=nil{
				// // 		nn, err := proom.TrackSendRtpPackets(pps.device.Streamname, "audio", b[:n]) //20*time.Millisecond)
				// // 		if err != nil || nn != n {
				// // 			log.Debug("audio forward error ", err, "->", nn)
				// // 			pps.bAudioStop = true
				// // 		}
				// // 	}
				// // }
				// time.Sleep(20 * time.Millisecond)
			}

			//	}()
		} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeH264) {
			//转发视频频数据包
			//go func() {
			b := make([]byte, 1500)
			go func() {
				ticker := time.NewTicker(time.Millisecond * 200)
				for range ticker.C {
					errSend := pps.pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
					if errSend != nil {
						log.Debug(errSend)
						break
					}
					if pps.bVideoStop {
						break
					}
				}
			}()
			pps.bVideoStop = false
			// bFirstStart := false
			// h264p := make([]byte, 1024*64)
			// h264len := 0
			for !pps.bVideoStop {
				// Read
				if stream.IsRtpStream() {
					n, _, readErr := track.Read(b)
					if readErr != nil {
						log.Debug(readErr)
						break
					}
					stream.SendStreamVideo(b[:n])
					// if n > 30 {
					// 	n = 30
					// }
					// log.InfoHex(b, n)
				} else {
					pkt, _, err := track.ReadRTP()
					if err != nil {
						log.Debug(err)
						break
						//return
					}
					// n, _, readErr := track.Read(b)
					// if readErr != nil {
					// 	log.Debug(readErr)
					// 	break
					// }
					// datas, bstart := UnpackRTP2H264(pkt.Payload)

					// if bstart && !bFirstStart {
					// 	bFirstStart = true
					// 	h264p = append(h264p[:h264len], datas...)
					// 	h264len += len(datas)
					// 	continue
					// } else if bstart && bFirstStart {
					// 	h264p = append(h264p[:h264len], datas...)
					// 	h264len += len(datas)
					// 	continue
					// } else if !bstart && bFirstStart {
					// 	bFirstStart = false
					// 	h264p = append(h264p[:h264len], datas...)
					// 	h264len += len(datas)
					// } else if !bstart && !bFirstStart {
					// 	h264p = append(h264p[:h264len], datas...)
					// 	h264len += len(datas)
					// }

					h264packet := H264Packet{}
					datas, err := h264packet.GetRTPRawH264(pkt)
					if err != nil {
						log.Debug(err)
						break
					}
					stream.SendStreamVideo(datas)
				}
				// ptklen := len(pkt.Payload)
				// log.Debug("send video rtp payload len ->", ptklen, "raw h264 len->", len(datas))
				// if ptklen >= 10 {
				// 	log.InfoHex(pkt.Payload, 10)
				// } else {
				// 	log.InfoHex(pkt.Payload, ptklen)
				// }
				// 	if h264p == nil {
				// 		continue
				// 	}
				// 	log.Debug("h264 data len", h264len)
				// 	stream.SendStreamVideo(h264p[:h264len])
				// 	if h264len > 10 {
				// 		h264len = 10
				// 	}
				// 	log.InfoHex(h264p, h264len)
				// 	h264len = 0
				// 	time.Sleep(time.Second / 30)
				// }
				// log.Debug(b[0:10])
				// for _, proom := range pps.device.PRoomList {
				// 	if proom != nil {
				// 		// proom.TrackPublished(device.Streamname)
				// 		// track.Read
				// 		// track.PayloadType()

				// 		nn, err := proom.TrackSendRtpPackets(pps.device.Streamname, "video", b[:n]) //time.Second/30)
				// 		if err != nil || nn != n {
				// 			log.Debug("video forward error ", err, "->", nn)
				// 			pps.bVideoStop = true
				// 		} else {
				// 			// log.Debug("video forward ok data len", n, "room is ", proom)
				// 		}
				// 	}
				// }

			}
			//}()
		}
	}
	return nil
}

func (pps *Stream) writeTrackToDisk(w media.Writer, track *webrtc.TrackRemote) error {
	for pps.ctx.Err() == nil {
		pkt, _, err := track.ReadRTP()
		// fmt.Println(filename, pkt.Timestamp)
		if err != nil {
			if pps.ctx.Err() != nil {
				return nil
			}
			log.Debug("writeTrackToDisk error ", err, w, track)
			return err
		}

		if w == nil {
			continue
		}

		if err := w.WriteRTP(pkt); err != nil {
			if len(pkt.Payload) <= 2 {
				continue
			}
			logrus.Warnf("Ignore write RTP %vB err %+v\n", len(pkt.Payload), err)
		} else {
			log.Debug("WriteRTP track pkt type->", pkt.PayloadType, "pkt", pkt)
		}
	}

	return pps.ctx.Err()
}

func (pps *Stream) Stop() bool {
	pps.cancel()
	if pps.hasAudioTrack {
		<-pps.audioFinish
	}
	if pps.hasVideoTrack {
		<-pps.videoFinish
	}

	if pps.hasVideoTrack && pps.hasAudioTrack {
		audiopath := pps.savePath + pps.Display + "_audio.ogg"
		videopath := pps.savePath + pps.Display + "_video.h264"

		cmd := exec.Command("ffmpeg",
			"-i",
			audiopath,
			"-i",
			videopath,
			pps.savePath+pps.Display+".ts",
			"-y")
		if err := cmd.Run(); err != nil {
			logrus.Errorf("拼接音频和视频失败:%v", err)
			return false
		}
		return true
	}
	return false
}
func apiRtcRequest(msg *Message, sdp *webrtc.SessionDescription, device *livekitclient.Device, time_out time.Duration) (string, error) {
	req := &Session{}

	req.DeviceId = device.Deviceid //msg.SeqID //config.Config.Mqtt.CLIENTID
	msg.SeqID = config.Config.Mqtt.CLIENTID
	msg.RtcSession = *sdp
	sdp_topic := TOPIC_OFFER
	msg.Topicprefix = config.Config.Mqtt.SUBTOPIC[:len(config.Config.Mqtt.SUBTOPIC)-2]
	if sdp.Type == webrtc.SDPTypeOffer {
		req.Type = CMDMSG_OFFER
		msg.Mode = MODE_OFFER
		sdp_topic = TOPIC_OFFER
	} else if sdp.Type == webrtc.SDPTypeAnswer {
		req.Type = CMDMSG_ANSWER
		msg.Mode = MODE_ANSWER
		sdp_topic = TOPIC_ANSWER
	}
	msg.Video = true //
	msg.Audio = true //
	// msg.SeqID = device.Deviceid
	msg.Describestreamname = device.Streamname //"testdevicestream"
	req.Data = enc.Encode(msg)
	sdpmsg := PublishMsg{
		WEB_SEQID: config.Config.Mqtt.CLIENTID, //msg.SeqID,
		Topic:     sdp_topic,
		Msg:       req,
		BTodevice: true,
	}
	log.Debug("sdp signale %s", msg.SeqID, "msg.Topicprefix", msg.Topicprefix)
	SendMsg(sdpmsg)
	if sdp.Type == webrtc.SDPTypeOffer {
		timeout := time.After(time_out)
		for {
			select {
			case <-timeout:
				log.Debugf("sdp wait timeout")
				return "", errors.New("timeout")
			case sdpanswer := <-SDPCh:
				log.Debugf("get sdpanswer", sdpanswer)
				return (*sdpanswer).RtcSession.SDP, nil

			}
		}
	} else {
		return "", nil
	}

}
func NewStream(host, room, display, savePath string, device *livekitclient.Device, msg *Message) (*Stream, error) {
	var err error
	pion_stream := &Stream{
		Host:          host,
		Room:          room,
		Display:       display,
		rtcUrl:        "webrtc://" + host + "/" + room + "/" + display,
		savePath:      savePath,
		hasAudioTrack: false,
		hasVideoTrack: false,
		videoFinish:   make(chan struct{}, 1),
		audioFinish:   make(chan struct{}, 1),
		device:        device,
	}
	pion_stream.ctx, pion_stream.cancel = context.WithCancel(context.Background())

	//创建PeerConncetion
	pion_stream.pc, err = newPeerConnection(webrtc.Configuration{
		ICEServers: (*msg).ICEServers,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "创建PeerConnection失败")
	}

	//设置方向
	pion_stream.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})
	pion_stream.pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	})

	//创建offer
	offer, err := pion_stream.pc.CreateOffer(nil)
	if err != nil {
		return nil, errors.Wrap(err, "创建Local offer失败")
	}

	// 设置本地sdp
	if err = pion_stream.pc.SetLocalDescription(offer); err != nil {
		return nil, errors.Wrap(err, "设置Local SDP失败")
	}
	// msg := Message{}
	// device := livekitclient.Device{
	// 	Deviceid: "device_1",
	// }
	// 设置远端SDP
	timeout := 10 * time.Second
	answer, err := apiRtcRequest(msg, &offer, device, timeout)
	if err != nil {
		return nil, errors.Wrap(err, "SDP协商失败")
	}

	if err = pion_stream.pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: answer,
	}); err != nil {
		return nil, errors.Wrap(err, "设置Remote SDP失败")
	}
	// if !bWrite {
	// 	for _, proom := range device.PRoomList {
	// 		if proom != nil {
	// 			proom.RTPTrackPublished(device.Streamname)
	// 		}
	// 	}
	// }
	var pstream *media_interface.Stream
	// bPublish := false
	pion_stream.pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		fmt.Println("------------------")

		// if !bWrite && bPublish && strings.Contains(track.Codec().MimeType, "video") {
		if !bWrite {
			pstream, err = publishStream(track, device.Deviceid, device.Streamname)
			if err != nil {
				log.Debug("publishStream fail", err)
				return
			}
			// bPublish = true
		}
		err = pion_stream.onTrack(track, receiver, pstream)
		if err != nil {
			codec := track.Codec()
			logrus.Errorf("Handle  track %v, pt=%v\nerr %v", codec.MimeType, codec.PayloadType, err)
			pion_stream.cancel()
		}
		pion_stream.pc.Close()
	})

	pion_stream.pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		logrus.Infof("ICE state %v", state)

		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			if pion_stream.ctx.Err() != nil {
				return
			}

			logrus.Warnf("Close for ICE state %v", state)
			pion_stream.cancel()
			pion_stream.pc.Close()
			m := media_interface.GetGlobalStreamM()
			err := m.DeleteStream(device.Streamname)
			if err != nil {
				log.Debug("DeleteStream error", err)
				//return err
			}
			// for _, proom := range device.PRoomList {
			// 	if proom != nil {
			// 		proom.TrackClose(device.Streamname)
			// 	}
			// }
		}
	})
	key := "webrtc://" + host + "/" + room
	Streams.Store(key, pion_stream)
	// timeout1 := time.After(20 * time.Minute)
	for pion_stream.ctx.Err() == nil {
		if pion_stream.bAudioStop && pion_stream.bVideoStop {
			break
		}
		//log.Debug("timeoutu->", timeoutu)
		//pion_stream.Stop()
		//return pion_stream, nil
		time.Sleep(time.Second)
	}

	// time.Sleep(time.Second)
	// }
	pion_stream.pc.Close()
	time.Sleep(time.Second)
	log.Debug("forward task end")
	return pion_stream, nil
}

func newPeerConnection(configuration webrtc.Configuration) (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	// for _, extension := range []string{sdp.SDESMidURI, sdp.SDESRTPStreamIDURI, sdp.TransportCCURI} {
	// 	if extension == sdp.TransportCCURI {
	// 		continue
	// 	}
	// 	if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
	// 		return nil, err
	// 	}
	// }
	// //RTC_CODEC_H264_PROFILE_42E01F_LEVEL_ASYMMETRY_ALLOWED_PACKETIZATION_MODE kvs used level 3.1
	// // https://github.com/pion/ion/issues/130
	// // https://github.com/pion/ion-sfu/pull/373/files#diff-6f42c5ac6f8192dd03e5a17e9d109e90cb76b1a4a7973be6ce44a89ffd1b5d18R73
	// for _, extension := range []string{sdp.SDESMidURI, sdp.SDESRTPStreamIDURI, sdp.AudioLevelURI} {
	// 	if extension == sdp.AudioLevelURI {
	// 		continue
	// 	}
	// 	if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeAudio); err != nil {
	// 		return nil, err
	// 	}
	// }

	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
	return api.NewPeerConnection(configuration)
}
