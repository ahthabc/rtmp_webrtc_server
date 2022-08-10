package livekitclient

import (
	"fmt"
	"strings"
	"time"

	// "github.com/livekit/server-sdk-go/pkg/media/ivfwriter"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"
	ionsdk "github.com/pion/ion-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264writer"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/xiangxud/rtmp_webrtc_server/identity"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	// "github.com/livekit/server-sdk-go/pkg/samplebuilder"
)

const (
// sid = "ion"
// uid = ionsdk.RandomKey(6)
)

func (t *LocalTrackPublication) saveToDisk(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	codec := track.Codec()
	var fileWriter media.Writer
	var err error
	if strings.EqualFold(codec.MimeType, webrtc.MimeTypeOpus) {
		log.Infof("Got Opus track, saving to disk as ogg (48 kHz, 2 channels)")
		fileWriter, err = oggwriter.New(fmt.Sprintf("%d_%d.ogg", codec.PayloadType, track.SSRC()), 48000, 2)
	} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeVP8) {
		log.Infof("Got VP8 track, saving to disk as ivf")
		fileWriter, err = ivfwriter.New(fmt.Sprintf("%d_%d.ivf", codec.PayloadType, track.SSRC()))
	} else if strings.EqualFold(codec.MimeType, webrtc.MimeTypeH264) {
		log.Infof("Got H264 track, saving to disk as h264")
		fileWriter, err = h264writer.New(fmt.Sprintf("%d_%d.h264", codec.PayloadType, track.SSRC()))
	}
	if err != nil {
		log.Errorf("error: %v", err)
		fileWriter.Close()
		return
	}

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			log.Warnf("track.ReadRTP error: %v", err)
			break
		}
		if err := fileWriter.WriteRTP(rtpPacket); err != nil {
			log.Warnf("fileWriter.WriteRTP error: %v", err)
			break
		}
	}
}
func (t *LocalTrackPublication) INORoomRTCJoin(r *Room, streamname, identify string) (*ionsdk.RTC, error) {

	// join room
	uid := streamname + ":" + identify
	err := r.IONRoom.Join(
		ionsdk.JoinInfo{
			Sid:         identify,
			Uid:         uid,
			DisplayName: uid,
			Role:        ionsdk.Role_Host,
			Protocol:    ionsdk.Protocol_WebRTC,
			Direction:   ionsdk.Peer_BILATERAL,
		},
	)

	if err != nil {
		log.Errorf("Join error: %v", err)
		return nil, err
	}
	// new sdk engine
	config := ionsdk.RTCConfig{
		WebRTC: ionsdk.WebRTCTransportConfig{
			VideoMime: ionsdk.MimeTypeH264,
		},
	}
	joinedch := make(chan struct{})
	r.IONRoom.OnJoin = func(success bool, info ionsdk.RoomInfo, err error) {
		// THIS IS ROOM SINGAL API
		// ===============================
		rtc, err1 := ionsdk.NewRTC(r.IONConnector, config)
		if err1 != nil {
			log.Error(err1)
			return
		}

		// user define receiving rtp
		rtc.OnTrack = t.saveToDisk
		rtc.GetPubTransport().GetPeerConnection().OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
			log.Infof("Connection state changed: %s", state)
		})
		rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
			log.Infof("dc: %v", dc.Label())
		}

		rtc.OnError = func(err error) {
			log.Errorf("err: %v", err)
		}
		configjoin := ionsdk.JoinConfig{}
		configjoin.SetNoPublish()
		configjoin.SetNoSubscribe()
		configjoin.SetNoAutoSubscribe()
		uid = ionsdk.RandomKey(6)
		err1 = rtc.Join(identify, uid, &configjoin)
		if err1 != nil {
			log.Errorf("error: %v", err1)
			return
		}
		log.Infof("rtc.Join ok sid=%v username =%v", identify, uid)

		// err = rtc.Join(session, ionsdk.RandomKey(4))
		// if err != nil {
		// 	log.Errorf("error: %v", err)
		// 	return
		// }
		t.IONRtc = rtc
		joinedch <- struct{}{}
	}
	<-joinedch
	return t.IONRtc, nil
}
func (r *Room) CreateIonRoom(addr, session string) (*ionsdk.Room, error) {
	log.Debug("CreateIonRoom: ", addr, "  session: ", session)
	r.HostIon = addr
	connector := ionsdk.NewConnector(addr)
	// uid := ionsdk.RandomKey(6)
	room := ionsdk.NewRoom(connector)
	peers := room.GetPeers(session)
	if len(peers) != 0 {
		log.Debug("room is exit peers :", peers)
		// err := room.Join(
		// 	ionsdk.JoinInfo{
		// 		Sid:         session,
		// 		Uid:         uid,
		// 		DisplayName: uid,
		// 		Role:        ionsdk.Role_Host,
		// 		Protocol:    ionsdk.Protocol_WebRTC,
		// 		Direction:   ionsdk.Peer_BILATERAL,
		// 	},
		// )

		// if err != nil {
		// 	log.Errorf("Join error: %v", err)
		// 	return nil, err
		// }
		r.IONRoom = room
		r.IONConnector = connector
		return room, nil
	}
	// THIS IS ROOM MANAGEMENT API
	// ==========================
	// create room
	err := room.CreateRoom(ionsdk.RoomInfo{Sid: session})
	if err != nil {
		log.Errorf("error:%v", err)
		return nil, err
	}

	// // new sdk engine
	// config := ionsdk.RTCConfig{
	// 	WebRTC: ionsdk.WebRTCTransportConfig{
	// 		VideoMime: ionsdk.MimeTypeH264,
	// 	},
	// }
	// // THIS IS ROOM SINGAL API
	// // ===============================
	// rtc, err := ionsdk.NewRTC(connector, config)
	// if err != nil {
	// 	log.Error(err)
	// 	// return err
	// }

	// // user define receiving rtp
	// rtc.OnTrack = r.saveToDisk

	// rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
	// 	log.Infof("dc: %v", dc.Label())
	// }

	// rtc.OnError = func(err error) {
	// 	log.Errorf("err: %v", err)
	// }
	// err = rtc.Join(session, ionsdk.RandomKey(4))
	// if err != nil {
	// 	log.Errorf("error: %v", err)
	// 	return nil, err
	// }
	// log.Infof("rtc.Join ok sid=%v", session)
	// // err = rtc.Join(session, ionsdk.RandomKey(4))
	// // if err != nil {
	// // 	log.Errorf("error: %v", err)
	// // 	return
	// // }
	// r.IONRtc = rtc
	room.OnJoin = func(success bool, info ionsdk.RoomInfo, err error) {
		log.Infof("OnJoin success = %v, info = %v, err = %v", success, info, err)

	}
	room.OnLeave = func(success bool, err error) {
		log.Infof("OnLeave success = %v err = %v", success, err)
	}

	room.OnPeerEvent = func(state ionsdk.PeerState, peer ionsdk.PeerInfo) {
		log.Infof("OnPeerEvent state = %v, peer = %v", state, peer)
	}

	room.OnMessage = func(from string, to string, data map[string]interface{}) {
		log.Infof("OnMessage from = %v, to = %v, data = %v", from, to, data)
	}

	room.OnDisconnect = func(sid, reason string) {
		log.Infof("OnDisconnect sid = %v, reason = %v", sid, reason)
	}

	room.OnRoomInfo = func(info ionsdk.RoomInfo) {
		log.Infof("OnRoomInfo info=%v", info)
	}

	// join room
	// err = room.Join(
	// 	ionsdk.JoinInfo{
	// 		Sid:         session,
	// 		Uid:         uid,
	// 		DisplayName: uid,
	// 		Role:        ionsdk.Role_Host,
	// 		Protocol:    ionsdk.Protocol_WebRTC,
	// 		Direction:   ionsdk.Peer_BILATERAL,
	// 	},
	// )

	// if err != nil {
	// 	log.Errorf("Join error: %v", err)
	// 	return nil, err
	// }
	r.IONRoom = room
	r.IONConnector = connector
	return room, nil
}

// func (t *LocalTrackPublication) ConnectRoomIon(host,  identity string) error {
// 	// host := "<host>"
// 	// apiKey := "api-key"
// 	// apiSecret := "api-secret"
// 	// roomName := "myroom"
// 	// identity := "botuser"
// 	room, err := lksdk.ConnectToRoom(host, lksdk.ConnectInfo{
// 		APIKey:              apikey,
// 		APISecret:           apisecret,
// 		RoomName:            roomname,
// 		ParticipantIdentity: identity,
// 	})
// 	if err != nil {
// 		log.Debug(err)
// 		return err
// 	}
// 	t.LiveKitRoomConnect = room
// 	room.Callback.OnTrackSubscribed = t.TrackSubscribed
// 	return nil
// 	// room.Disconnect()
// }
func (r *Room) TrackPublished_to_ION(streamname string) error {
	// - `in` implements io.ReadCloser, such as buffer or file
	// - `mime` has to be one of webrtc.MimeType...
	// videoTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264})
	r.ionlock.Lock()
	defer r.ionlock.Unlock()
	if r.Ctx != nil && r.IONRoom == nil {
		var err error
		sn, _ := identity.GetSN()
		r.IONRoom, err = r.CreateIonRoom(r.HostIon, sn)
		if err != nil {
			log.Debug("room->", sn, "create room ok", r)
			return err
		}

	}
	t := r.Localtracks[streamname]
	if t == nil {
		t = &LocalTrackPublication{Streamname: streamname}
		t.INORoomRTCJoin(r, streamname, r.Identity)
		log.Debug("ion track->", streamname, "<-is nil ,Connect room", t, r)
	} else {
		if !t.IONRtc.Connected() {
			t.INORoomRTCJoin(r, streamname, r.Identity)
			log.Debug("ion track->", streamname, "<-is nil ,re Connect room", t, r)
		}
	}
	if t.IONSfuTrack.VideoTrack == nil && t.Videopub == nil && t.IONSfuTrack.AudioTrack == nil && t.Audiopub == nil {
		videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, streamname+"-video", streamname)
		if err != nil {
			panic(err)
		}
		// r.RoomClient.MutePublishedTrack(r.Ctx,)
		// var local_video *lksdk.LocalTrackPublication
		if _, err = t.IONRtc.Publish(videoTrack); err != nil {
			log.Debug("Error publishing video track->", err)
			return err
		}
		t.IONSfuTrack.VideoTrack = videoTrack
		// r.Localtracks[streamname] = &LocalTrackPublication{p: local_video, Track: videoTrack, Trackname: streamname + "-video"}
		log.Debug("[TrackPublished_to_ION]", "published video track -> ", streamname)

		if t.IONSfuTrack.AudioTrack == nil {
			audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, streamname+"-audio", streamname)
			//audioTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
			if err != nil {
				panic(err)
			}
			// var local_audio *lksdk.LocalTrackPublication

			if _, err = t.IONRtc.Publish(audioTrack); err != nil {
				log.Debug("Error publishing audio track->", err)
				return err
			}
			t.IONSfuTrack.AudioTrack = audioTrack
			log.Debug("[TrackPublished_to_ION]", "published audio track -> ", streamname)
		}

		r.Localtracks[streamname] = t

	} else {
		log.Debug(streamname, "is exit publish")
	}
	return nil
}
func (r *Room) RTPTrackPublished_to_ION(trackRemote []*webrtc.TrackRemote, streamname string) error {
	// - `in` implements io.ReadCloser, such as buffer or file
	// - `mime` has to be one of webrtc.MimeType...
	// videoTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264})
	// r.mux.lock()
	r.ionlock.Lock()
	defer r.ionlock.Unlock()
	if r.Ctx != nil && r.IONRoom == nil {
		var err error
		sn, _ := identity.GetSN()
		r.IONRoom, err = r.CreateIonRoom(r.HostIon, sn)
		if err != nil {
			log.Debug("room->", sn, "create room ok", r)
			return err
		}
	}
	t := r.Localtracks[streamname]
	if t == nil {
		t = &LocalTrackPublication{Streamname: streamname}
		t.INORoomRTCJoin(r, streamname, r.Identity)
		log.Debug("track->", streamname, "<-is nil ,Connect room", t, r)
	} else {
		if !t.IONRtc.Connected() {
			t.INORoomRTCJoin(r, streamname, r.Identity)
			log.Debug("ion track->", streamname, "<-is nil ,re Connect room", t, r)
		}
	}
	for _, v := range trackRemote {
		if v.Kind().String() == "video" {
			if t.IONSfuTrack.VideoRTPTrack == nil {
				if strings.Contains(v.Codec().MimeType, "video") {
					videoRTPTrack, err := webrtc.NewTrackLocalStaticRTP(v.Codec().RTPCodecCapability, streamname+"-video", streamname)
					if err != nil {
						panic(err)
					}
					var pub []*webrtc.RTPSender
					log.Debug("ion track publish", videoRTPTrack)
					if pub, err = t.IONRtc.Publish(videoRTPTrack); err != nil {
						log.Debug("Error publishing video RTP track->", err)
						return err
					}
					t.IONVideopub = pub[0]
					t.IONSfuTrack.VideoRTPTrack = videoRTPTrack
					r.Localtracks[streamname] = t
					log.Debug("[RTPTrackPublished_to_ION]", "published video track -> ", streamname)
				}
			}
		} else {
			if v.Kind().String() == "audio" {
				if t.IONSfuTrack.AudioRTPTrack == nil {
					if strings.Contains(v.Codec().MimeType, "audio") {
						audioRTPTrack, err := webrtc.NewTrackLocalStaticRTP(v.Codec().RTPCodecCapability, streamname+"-audio", streamname)
						if err != nil {
							panic(err)
						}
						// var local_audio *lksdk.LocalTrackPublication
						var pub []*webrtc.RTPSender
						log.Debug("ion track publish", audioRTPTrack)
						if pub, err = t.IONRtc.Publish(audioRTPTrack); err != nil {
							log.Debug("Error publishing audio track", err)
							return err
						}
						t.IONAudiopub = pub[0]
						t.IONSfuTrack.AudioRTPTrack = audioRTPTrack
						r.Localtracks[streamname] = t
						log.Debug("[RTPTrackPublished_to_ION]", "published audio track -> ", streamname)
					}
				}
			}
		}
	}
	// r.Localtracks[streamname] = t

	// } else {
	// 	log.Debug(streamname, "is exit publish")
	// }
	return nil
}
func (r *Room) TrackSendIonRtpPackets(trackname, kind string, data []byte) (n int, err error) {
	if trackname == "" {
		log.Debug("Track name is null")
		return 0, fmt.Errorf("input trackname is null")
	}
	// var t *webrtc.TrackLocalStaticSample
	var t *webrtc.TrackLocalStaticRTP
	track := r.Localtracks[trackname]
	if track == nil {
		log.Debug("TrackSendIonRtpPackets:", "Track is nil ->", trackname, "<- no to publish")
		return 0, fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		t = track.IONSfuTrack.VideoRTPTrack
	} else if kind == "audio" {
		t = track.IONSfuTrack.AudioRTPTrack
	}
	if t == nil {
		log.Debug("TrackSendIonRtpPackets:", "t is nil ->", trackname, "<- no to publish")
		return 0, fmt.Errorf(" track is null,no to publish")
	}
	var sb *samplebuilder.SampleBuilder
	packets := &rtp.Packet{}
	if err := packets.Unmarshal(data); err != nil {
		return 0, err
	}
	sb.Push(packets)
	for _, p := range sb.PopPackets() {
		err = t.WriteRTP(p)
		if err != nil {
			log.Debug("[TrackSendIonRtpPackets] error", err)
			return 0, err
		}
	}
	//n, err = t.Write(data)
	return len(data), nil

}
func (r *Room) TrackSendIonData(trackname, kind string, data []byte, duration time.Duration) error {
	if trackname == "" {
		log.Debug("Track name is null")
		return fmt.Errorf("input trackname is null")
	}
	var t *webrtc.TrackLocalStaticSample
	track := r.Localtracks[trackname]
	if track == nil {
		log.Debug("Track is nil ->", trackname, "<- no to publish")
		return fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		t = track.IONSfuTrack.VideoTrack
	} else if kind == "audio" {
		t = track.IONSfuTrack.AudioTrack
	}
	if t == nil {
		log.Debug("Track is nil ->", trackname, "<- no to publish")
		return fmt.Errorf(" track is null,no to publish")
	}

	if videoErr := t.WriteSample(media.Sample{
		Data:     data,
		Duration: duration,
	}); videoErr != nil {
		log.Debug("WriteSample err", videoErr)
		// r.ConnectRoom()
		return nil //fmt.Errorf("WriteSample err %s", vedioErr)
	}

	return nil
}
func (r *Room) TrackCloseION(streamname string) error {
	if t := r.Localtracks[streamname]; t != nil {
		var pub []*webrtc.RTPSender
		if r.Localtracks[streamname].IONVideopub != nil {
			pub = append(pub, r.Localtracks[streamname].IONVideopub)
		}
		if r.Localtracks[streamname].IONAudiopub != nil {
			pub = append(pub, r.Localtracks[streamname].IONAudiopub)
		}
		if pub != nil {
			t.IONRtc.UnPublish(pub...)
			t.LiveKitRoomConnect.Disconnect()
		}
		r.Localtracks[streamname] = nil
		log.Debug("track ", streamname, "lost ,now removed", r)
		//r.RoomConnect.LocalParticipant.UnpublishTrack(r.RoomConnect.LocalParticipant.SID())
		// r.Localtracks[streamname+"-video"]
	}

	return nil
}
