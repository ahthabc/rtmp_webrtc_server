package livekitclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/livekit/server-sdk-go/pkg/samplebuilder"

	// "github.com/livekit/server-sdk-go/pkg/media/ivfwriter"
	// "github.com/livekit/server-sdk-go/pkg/samplebuilder"
	ionsdk "github.com/pion/ion-sdk-go"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264writer"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"

	// "github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/identity"
	"github.com/xiangxud/rtmp_webrtc_server/log"
)

type SfuTrack struct {
	VideoRTPTrack *webrtc.TrackLocalStaticRTP
	AudioRTPTrack *webrtc.TrackLocalStaticRTP
	VideoTrack    *webrtc.TrackLocalStaticSample
	AudioTrack    *webrtc.TrackLocalStaticSample
}
type LocalTrackPublication struct {
	LiveKitRoomConnect *lksdk.Room
	IONRtc             *ionsdk.RTC
	// IONRoomConnect     *ionsdk.Room
	Videopub        *lksdk.LocalTrackPublication
	Audiopub        *lksdk.LocalTrackPublication
	IONVideopub     *webrtc.RTPSender
	IONAudiopub     *webrtc.RTPSender
	LiveKitSfuTrack SfuTrack
	IONSfuTrack     SfuTrack
	livekitsb       *samplebuilder.SampleBuilder
	// publication *lksdk.RemoteTrackPublication
	// pliWriter       lksdk.PLIWriter
	// VideoRTPTrack *webrtc.TrackLocalStaticRTP
	// AudioRTPTrack *webrtc.TrackLocalStaticRTP
	// VideoTrack    *webrtc.TrackLocalStaticSample
	// AudioTrack    *webrtc.TrackLocalStaticSample
	// RemoteSDPOffer webrtc.SessionDescription `json:"sdpoffer"`
	// LocalSDPanswer webrtc.SessionDescription `json:"answer"`
	Streamname string
	// Trackname string
}
type Room struct {
	Token
	Ctx          context.Context
	RoomClient   *lksdk.RoomServiceClient
	LiveKitRoom  *livekit.Room
	livekitlock  sync.Mutex
	ionlock      sync.Mutex
	IONRoom      *ionsdk.Room
	IONConnector *ionsdk.Connector
	// IONRtc       *ionsdk.RTC
	Localtracks map[string]*LocalTrackPublication
}

func NewRoom(ctx context.Context, token *Token) *Room { //host, apiKey, apiSecret, roomName, identity string) *Room {

	return &Room{
		Ctx:         ctx,
		Token:       *token,
		Localtracks: make(map[string]*LocalTrackPublication),
	}
}

func (r *Room) CreateliveKitRoom(roomName string) (*Room, error) {
	var err error
	r.RoomClient = lksdk.NewRoomServiceClient(r.HostLiveKit, r.ApiKey, r.ApiSecret)
	r.RoomName = roomName
	// create a new room
	if r.Ctx != nil {
		r.LiveKitRoom, err = r.RoomClient.CreateRoom(r.Ctx, &livekit.CreateRoomRequest{
			Name: roomName,
		})
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return nil, fmt.Errorf("context is invalid")
}
func (t *LocalTrackPublication) ConnectRoom(host, apikey, apisecret, roomname, identity string) error {
	// host := "<host>"
	// apiKey := "api-key"
	// apiSecret := "api-secret"
	// roomName := "myroom"
	// identity := "botuser"
	room, err := lksdk.ConnectToRoom(host, lksdk.ConnectInfo{
		APIKey:              apikey,
		APISecret:           apisecret,
		RoomName:            roomname,
		ParticipantIdentity: identity,
	}, &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: t.TrackSubscribed,
		},
	})
	if err != nil {
		panic(err)
	}
	// room, err := lksdk.ConnectToRoom(host, lksdk.ConnectInfo{
	// 	APIKey:              apikey,
	// 	APISecret:           apisecret,
	// 	RoomName:            roomname,
	// 	ParticipantIdentity: identity,
	// })
	// if err != nil {
	// 	log.Debug(err)
	// 	return err
	// }
	t.LiveKitRoomConnect = room
	// room.Callback.OnTrackSubscribed = t.TrackSubscribed
	return nil
	// room.Disconnect()
}

// func (t *LocalTrackPublication) SetOffer(offer webrtc.SessionDescription) {
// 	t.RemoteSDPOffer = offer
// }
// func (t *LocalTrackPublication) SetAnswer(answer webrtc.SessionDescription) {
// 	t.LocalSDPanswer = answer
// }

// func (t *Room) ConnectRoom(streamname string) error {
// 	// host := "<host>"
// 	// apiKey := "api-key"
// 	// apiSecret := "api-secret"
// 	// roomName := "myroom"
// 	// identity := "botuser"
// 	room, err := lksdk.ConnectToRoom(r.Host, lksdk.ConnectInfo{
// 		APIKey:              r.ApiKey,
// 		APISecret:           r.ApiSecret,
// 		RoomName:            r.RoomName,
// 		ParticipantIdentity: streamname,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	room.Callback.OnTrackSubscribed = r.TrackSubscribed
// 	if t := r.Localtracks[streamname]; t != nil {
// 		t.RoomConnect = room
// 	}
// 	return nil
// 	// room.Disconnect()
// }
func (t *LocalTrackPublication) TrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {

	// }
	// func onTrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
	fileName := fmt.Sprintf("%s-%s", rp.Identity(), track.ID())
	fmt.Println("write track to file ", fileName)
	NewTrackWriter(track, rp.WritePLI, fileName)
	// t.pliWriter=
}

const (
	maxVideoLate = 1000 // nearly 2s for fhd video
	maxAudioLate = 200  // 4s for audio
)

type TrackWriter struct {
	sb     *samplebuilder.SampleBuilder
	writer media.Writer
	track  *webrtc.TrackRemote
}

func NewTrackWriter(track *webrtc.TrackRemote, pliWriter lksdk.PLIWriter, fileName string) (*TrackWriter, error) {
	var (
		sb     *samplebuilder.SampleBuilder
		writer media.Writer
		err    error
	)
	switch {
	case strings.EqualFold(track.Codec().MimeType, "video/vp8"):
		sb = samplebuilder.New(maxVideoLate, &codecs.VP8Packet{}, track.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
			pliWriter(track.SSRC())
		}))
		// ivfwriter use frame count as PTS, that might cause video played in a incorrect framerate(fast or slow)
		writer, err = ivfwriter.New(fileName + ".ivf")

	case strings.EqualFold(track.Codec().MimeType, "video/h264"):
		sb = samplebuilder.New(maxVideoLate, &codecs.H264Packet{}, track.Codec().ClockRate, samplebuilder.WithPacketDroppedHandler(func() {
			pliWriter(track.SSRC())
		}))
		writer, err = h264writer.New(fileName + ".h264")

	case strings.EqualFold(track.Codec().MimeType, "audio/opus"):
		sb = samplebuilder.New(maxAudioLate, &codecs.OpusPacket{}, track.Codec().ClockRate)
		writer, err = oggwriter.New(fileName+".ogg", 48000, track.Codec().Channels)

	default:
		return nil, errors.New("unsupported codec type")
	}

	if err != nil {
		return nil, err
	}

	t := &TrackWriter{
		sb:     sb,
		writer: writer,
		track:  track,
	}
	go t.start()
	return t, nil
}
func (t *TrackWriter) start() {
	defer t.writer.Close()
	for {
		pkt, _, err := t.track.ReadRTP()
		if err != nil {
			break
		}
		t.sb.Push(pkt)

		for _, p := range t.sb.PopPackets() {
			t.writer.WriteRTP(p)
		}
	}
}

func (r *Room) TrackSendLivekitRtpPackets(trackname, kind string, data []byte) (n int, err error) {
	if trackname == "" {
		log.Debug("Track name is null")
		return 0, fmt.Errorf("input trackname is null")
	}
	// var t *webrtc.TrackLocalStaticSample
	var t *webrtc.TrackLocalStaticRTP
	track := r.Localtracks[trackname]
	if track == nil {
		log.Debug("TrackSendLivekitRtpPackets: ", "Track is nil ->", trackname, "<- no to publish")
		return 0, fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		t = track.LiveKitSfuTrack.VideoRTPTrack
	} else if kind == "audio" {
		t = track.LiveKitSfuTrack.AudioRTPTrack
	}
	if t == nil {
		log.Debug("TrackSendLivekitRtpPackets: ", "t is nil ->", trackname, "<- no to publish")
		return 0, fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		packets := &rtp.Packet{}
		if err := packets.Unmarshal(data); err != nil {
			return 0, err
		}
		track.livekitsb.Push(packets)
		for _, p := range track.livekitsb.PopPackets() {
			err = t.WriteRTP(p)
			if err != nil {
				log.Debug("[TrackSendIonRtpPackets] error", err)
				return 0, err
			}
		}
		//n, err = t.Write(data)
		return len(data), nil
	} else {
		n, err = t.Write(data)
		return n, err
	}
}
func (r *Room) TrackSendLivekitData(trackname, kind string, data []byte, duration time.Duration) error {
	if trackname == "" {
		log.Debug("Track name is null")
		return fmt.Errorf("input trackname is null")
	}
	var t *webrtc.TrackLocalStaticSample
	track := r.Localtracks[trackname]
	if track == nil {
		log.Debug("TrackSendLivekitData:", "Track is nil ->", trackname, "<- no to publish")
		return fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		t = track.LiveKitSfuTrack.VideoTrack
	} else if kind == "audio" {
		t = track.LiveKitSfuTrack.AudioTrack
	}
	if t == nil {
		log.Debug("TrackSendLivekitData: ", "t is nil ->", trackname, "<- no to publish")
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
func (r *Room) TrackClose(streamname string) error {

	if t := r.Localtracks[streamname]; t != nil {
		if t.IONRtc != nil {
			t.IONRtc.Close()
		}
		if t.LiveKitRoomConnect == nil || t.LiveKitRoomConnect.LocalParticipant == nil {
			return nil
		}
		t.LiveKitRoomConnect.LocalParticipant.UnpublishTrack(t.LiveKitRoomConnect.LocalParticipant.SID())
		t.LiveKitRoomConnect.Disconnect()
		r.Localtracks[streamname] = nil
		log.Debug("track ", streamname, "lost ,now removed", r)
		//r.RoomConnect.LocalParticipant.UnpublishTrack(r.RoomConnect.LocalParticipant.SID())
		// r.Localtracks[streamname+"-video"]
	}

	return nil
}
func (r *Room) TrackPublished(streamname string) error {
	// - `in` implements io.ReadCloser, such as buffer or file
	// - `mime` has to be one of webrtc.MimeType...
	// videoTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264})
	r.livekitlock.Lock()
	defer r.livekitlock.Unlock()
	if r.Ctx != nil && r.LiveKitRoom == nil {
		var err error
		sn, _ := identity.GetSN()
		r.LiveKitRoom, err = r.RoomClient.CreateRoom(r.Ctx, &livekit.CreateRoomRequest{
			Name: sn,
		})
		if err != nil {
			log.Debug("room->", sn, "create room ok", r)
			return err
		}
	}
	t := r.Localtracks[streamname]
	if t == nil {
		t = &LocalTrackPublication{Streamname: streamname}
		t.ConnectRoom(r.HostLiveKit, r.ApiKey, r.ApiSecret, r.RoomName, streamname+":"+r.Identity)
		log.Debug("track->", streamname, "<-is nil ,Connect room", t, r)
	} else {
		if t.LiveKitRoomConnect == nil {
			t.ConnectRoom(r.HostLiveKit, r.ApiKey, r.ApiSecret, r.RoomName, streamname+":"+r.Identity)
			log.Debug("track->", streamname, "<-is nil ,re Connect room", t, r)
		}
	}
	if t.LiveKitSfuTrack.VideoTrack == nil && t.Videopub == nil && t.LiveKitSfuTrack.AudioTrack == nil && t.Audiopub == nil {
		videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, streamname+"-video", streamname)
		if err != nil {
			panic(err)
		}
		// r.RoomClient.MutePublishedTrack(r.Ctx,)
		// var local_video *lksdk.LocalTrackPublication
		if t.Videopub, err = t.LiveKitRoomConnect.LocalParticipant.PublishTrack(videoTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-video"}); err != nil {
			log.Debug("Error publishing video track->", err)
			return err
		}
		t.LiveKitSfuTrack.VideoTrack = videoTrack
		// r.Localtracks[streamname] = &LocalTrackPublication{p: local_video, Track: videoTrack, Trackname: streamname + "-video"}
		log.Debug("[TrackPublished]", "published video track -> ", streamname)

		if t.LiveKitSfuTrack.AudioTrack == nil {
			audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, streamname+"-audio", streamname)
			//audioTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
			if err != nil {
				panic(err)
			}
			// var local_audio *lksdk.LocalTrackPublication

			if t.Audiopub, err = t.LiveKitRoomConnect.LocalParticipant.PublishTrack(audioTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-audio"}); err != nil {
				log.Debug("Error publishing audio track->", err)
				return err
			}
			t.LiveKitSfuTrack.AudioTrack = audioTrack
			log.Debug("[TrackPublished]", "published audio track -> ", streamname)
		}

		r.Localtracks[streamname] = t

	} else {
		log.Debug(streamname, "is exit publish")
	}
	return nil
}
func (r *Room) RTPTrackPublished(trackRemote *webrtc.TrackRemote, streamname string) error {
	// - `in` implements io.ReadCloser, such as buffer or file
	// - `mime` has to be one of webrtc.MimeType...
	// videoTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264})
	// r.mux.lock()
	r.livekitlock.Lock()
	defer r.livekitlock.Unlock()
	if r.Ctx != nil && r.LiveKitRoom == nil {
		var err error
		sn, _ := identity.GetSN()
		r.LiveKitRoom, err = r.RoomClient.CreateRoom(r.Ctx, &livekit.CreateRoomRequest{
			Name: sn,
		})
		if err != nil {
			log.Debug("room->", sn, "create room ok", r)
			return err
		}
	}
	t := r.Localtracks[streamname]
	if t == nil {
		t = &LocalTrackPublication{Streamname: streamname}
		t.ConnectRoom(r.HostLiveKit, r.ApiKey, r.ApiSecret, r.RoomName, streamname+":"+r.Identity)
		log.Debug("track->", streamname, "<-is nil ,Connect room", t, r)
	} else {
		if t.LiveKitRoomConnect == nil {
			t.ConnectRoom(r.HostLiveKit, r.ApiKey, r.ApiSecret, r.RoomName, streamname+":"+r.Identity)
			log.Debug("track->", streamname, "<-is nil ,re Connect room", t, r)
		}

	}

	if t.LiveKitSfuTrack.VideoRTPTrack == nil {
		if strings.Contains(trackRemote.Codec().MimeType, "video") {
			if t.livekitsb == nil {
				t.livekitsb = samplebuilder.New(maxVideoLate, &codecs.H264Packet{}, trackRemote.Codec().ClockRate) //, samplebuilder.WithPacketDroppedHandler(func() {
				// 	t.Videopub.WritePLI(trackRemote.SSRC())
				// }))
				// t.Videopub.WritePLI
			}
			videoRTPTrack, err := webrtc.NewTrackLocalStaticRTP(trackRemote.Codec().RTPCodecCapability, streamname+"-video", streamname)
			if err != nil {
				panic(err)
			}

			// r.RoomClient.MutePublishedTrack(r.Ctx,)
			// var local_video *lksdk.LocalTrackPublication
			if t.LiveKitRoomConnect != nil && t.LiveKitRoomConnect.LocalParticipant != nil {
				if t.Videopub, err = t.LiveKitRoomConnect.LocalParticipant.PublishTrack(videoRTPTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-video"}); err != nil {
					log.Debug("Error publishing video RTP track->", err)
					return err
				}
				t.LiveKitSfuTrack.VideoRTPTrack = videoRTPTrack
				r.Localtracks[streamname] = t
				// r.Localtracks[streamname] = &LocalTrackPublication{p: local_video, Track: videoTrack, Trackname: streamname + "-video"}
				log.Debug("[RTPTrackPublished]", "published video track -> ", streamname)
			}
		}
	}
	if t.LiveKitSfuTrack.AudioRTPTrack == nil {
		if strings.Contains(trackRemote.Codec().MimeType, "audio") {
			audioRTPTrack, err := webrtc.NewTrackLocalStaticRTP(trackRemote.Codec().RTPCodecCapability, streamname+"-audio", streamname)
			//audioTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
			if err != nil {
				panic(err)
			}
			// var local_audio *lksdk.LocalTrackPublication
			if t.LiveKitRoomConnect != nil && t.LiveKitRoomConnect.LocalParticipant != nil {
				if t.Audiopub, err = t.LiveKitRoomConnect.LocalParticipant.PublishTrack(audioRTPTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-audio"}); err != nil {
					log.Debug("Error publishing audio track", err)
					return err
				}
				t.LiveKitSfuTrack.AudioRTPTrack = audioRTPTrack
				r.Localtracks[streamname] = t
				log.Debug("[RTPTrackPublished]", "published audio track -> ", streamname)
			}
		}
	}

	// r.Localtracks[streamname] = t

	// } else {
	// 	log.Debug(streamname, "is exit publish")
	// }
	return nil
}
func (r *Room) Close() {
	for _, t := range r.Localtracks {
		if t == nil {
			continue
		}
		if t.LiveKitRoomConnect != nil {
			t.LiveKitRoomConnect.Disconnect()
		}
	}
	if r.IONRoom != nil {
		r.IONRoom.Close()
	}
}
