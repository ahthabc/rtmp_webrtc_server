package livekitclient

import (
	"context"
	"fmt"
	"time"

	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"

	// "github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/identity"
	"github.com/xiangxud/rtmp_webrtc_server/log"
)

type LocalTrackPublication struct {
	RoomConnect *lksdk.Room
	Videopub    *lksdk.LocalTrackPublication
	Audiopub    *lksdk.LocalTrackPublication
	VideoTrack  *webrtc.TrackLocalStaticSample
	AudioTrack  *webrtc.TrackLocalStaticSample
	Streamname  string
	// Trackname string
}
type Room struct {
	Token
	Ctx        context.Context
	RoomClient *lksdk.RoomServiceClient
	Room       *livekit.Room
	// RoomConnect *lksdk.Room
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
	r.RoomClient = lksdk.NewRoomServiceClient(r.Host, r.ApiKey, r.ApiSecret)
	r.RoomName = roomName
	// create a new room
	if r.Ctx != nil {
		r.Room, err = r.RoomClient.CreateRoom(r.Ctx, &livekit.CreateRoomRequest{
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
	})
	if err != nil {
		panic(err)
	}
	t.RoomConnect = room

	return nil
	// room.Disconnect()
}

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
func (r *Room) TrackSubscribed(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {

}
func (r *Room) TrackSendData(trackname, kind string, data []byte, duration time.Duration) error {
	if trackname == "" {
		log.Debug("Track name is null")
		return fmt.Errorf("input trackname is null")
	}
	var t *webrtc.TrackLocalStaticSample
	track := r.Localtracks[trackname]
	if track == nil {
		log.Debug("Track is nil ->", track.Streamname, "<- no to publish")
		return fmt.Errorf(" track is null,no to publish")
	}
	if kind == "video" {
		t = track.VideoTrack
	} else if kind == "audio" {
		t = track.AudioTrack
	}
	if t == nil {
		log.Debug("Track is nil ->", track.Streamname, "<- no to publish")
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
		t.RoomConnect.LocalParticipant.UnpublishTrack(t.RoomConnect.LocalParticipant.SID())
		t.RoomConnect.Disconnect()
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

	if r.Ctx != nil && r.Room == nil {
		var err error
		sn, _ := identity.GetSN()
		r.Room, err = r.RoomClient.CreateRoom(r.Ctx, &livekit.CreateRoomRequest{
			Name: "RTMP-" + sn,
		})
		if err != nil {
			log.Debug("room->", "RTMP-"+sn, "create room ok", r)
			return err
		}
	}
	t := r.Localtracks[streamname]
	if t == nil {
		t = &LocalTrackPublication{Streamname: streamname}
		t.ConnectRoom(r.Host, r.ApiKey, r.ApiSecret, r.RoomName, streamname+":"+r.Identity)
		log.Debug("track->", streamname, "<-is nil ,Connect room", t)
	}
	if t.VideoTrack == nil {
		videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, streamname+"-video", streamname)
		if err != nil {
			panic(err)
		}
		// var local_video *lksdk.LocalTrackPublication
		if t.Videopub, err = t.RoomConnect.LocalParticipant.PublishTrack(videoTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-video"}); err != nil {
			log.Debug("Error publishing video track", err)
			return err
		}
		t.VideoTrack = videoTrack
		// r.Localtracks[streamname] = &LocalTrackPublication{p: local_video, Track: videoTrack, Trackname: streamname + "-video"}
		log.Debug("published video track -> ", streamname)
		if t.AudioTrack == nil {
			audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, streamname+"-audio", streamname)
			//audioTrack, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus})
			if err != nil {
				panic(err)
			}
			// var local_audio *lksdk.LocalTrackPublication

			if t.Audiopub, err = t.RoomConnect.LocalParticipant.PublishTrack(audioTrack, &lksdk.TrackPublicationOptions{Name: streamname + "-audio"}); err != nil {
				log.Debug("Error publishing audio track", err)
				return err
			}
			t.AudioTrack = audioTrack
			log.Debug("published audio track -> ", streamname)
		}

		r.Localtracks[streamname] = t

	} else {
		log.Debug(streamname, "is exit publish")
	}
	return nil
}
func (r *Room) Close() {
	for _, t := range r.Localtracks {
		t.RoomConnect.Disconnect()
	}
}
