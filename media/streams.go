package media

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Glimesh/go-fdkaac/fdkaac"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/livekitclient"

	// "github.com/xiangxud/rtmp_webrtc_server/livekitclient"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	opus "github.com/xiangxud/rtmp_webrtc_server/opus"
)

type StreamState byte

const (
	PEER_INIT StreamState = iota
	PEER_CONNECT
	PEER_TIMEOUT
	PEER_CLOSED
	PEER_DEADLINE
)
const (
	STREAM_INIT StreamState = iota
	STREAM_CONNECT
	STREAM_TIMEOUT
	STREAM_CLOSED
	STREAM_DEADLINE
)
const (
	USE_SFU_ION     = true
	USE_SFU_LIVEKIT = false
)

type peerInterface interface {
	InitPeer(peerid int64,
		peername string,
		username string,
		password string) error
	AddConnect(*webrtc.PeerConnection)
	AddAudioTrack(*webrtc.TrackLocalStaticSample)
	AddVideoTrack(*webrtc.TrackLocalStaticSample)
	AddAudioRTPTrack(*webrtc.TrackLocalStaticRTP)
	AddVideoRTPTrack(*webrtc.TrackLocalStaticRTP)
	SendPeerAudio([]byte) error
	SendPeerVideo([]byte) error
	VerifyPeer()
}

//webrtc 客户
type Peer struct {
	peerId                       string
	peerName                     string
	userName                     string
	streamName                   string
	passWord                     string
	status                       StreamState
	startTime                    time.Time
	endTime                      time.Time
	peerConnection               *webrtc.PeerConnection
	videoTrack, audioTrack       *webrtc.TrackLocalStaticSample
	videoRTPTrack, audioRTPTrack *webrtc.TrackLocalStaticRTP

	peerInterface
}

func (p *Peer) InitPeer(peerid string,
	peername string,
	username string,
	password string) {
	p.peerId = peerid
	p.peerName = peername
	p.userName = username
	p.passWord = password
	p.status = PEER_INIT
}
func (p *Peer) AddConnect(streamname string, pconn *webrtc.PeerConnection) {
	p.peerConnection = pconn
	p.startTime = time.Now()
	p.streamName = streamname
	p.status = PEER_CONNECT

}
func (p *Peer) AddAudioTrack(track *webrtc.TrackLocalStaticSample) {
	p.audioTrack = track
}
func (p *Peer) AddVideoTrack(track *webrtc.TrackLocalStaticSample) {
	p.videoTrack = track
}
func (p *Peer) AddAudioRTPTrack(track *webrtc.TrackLocalStaticRTP) {
	p.audioRTPTrack = track
}
func (p *Peer) AddVideoRTPTrack(track *webrtc.TrackLocalStaticRTP) {
	p.videoRTPTrack = track
}
func (p *Peer) SendPeerAudio(audiodata []byte) error {
	if p.status != PEER_CONNECT {
		// return fmt.Errorf("peer is not connected,status %d", p.status)
		// log.Debug("SendPeerAudio peer ", p.peerName, " is not connected,status ", p.status)
		return nil
	}
	if p.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
		p.status = PEER_CLOSED
		p.endTime = time.Now()
		log.Debug("SendPeerAudio peer ", p.peerName, " is closed,status ", p.status)
		return nil
	}
	if config.Config.Stream.Debug {
		log.Debug("SendPeerAudio peer name:", p.peerName, "stream name:", p.streamName, "opus len:", len(audiodata))
	}
	if audioErr := p.audioTrack.WriteSample(media.Sample{
		Data:     audiodata,
		Duration: 20 * time.Millisecond,
	}); audioErr != nil {
		log.Debug("WriteSample err", audioErr)
		return fmt.Errorf("WriteSample err %s", audioErr)
	}
	return nil
}
func (p *Peer) SendPeerVideo(videodata []byte) error {
	if p.status != PEER_CONNECT {
		// return fmt.Errorf("peer is not connected,status %d", p.status)
		// log.Debug("SendPeerVideo peer ", p.peerName, " is not connected,status ", p.status)
		return nil
	}
	if p.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
		p.status = PEER_CLOSED
		p.endTime = time.Now()
		// return fmt.Errorf("peer is closed,status %d", p.status)
		log.Debug("SendPeerVideo peer ", p.peerName, " is closed,status ", p.status)
		return nil
	}

	if videoErr := p.videoTrack.WriteSample(media.Sample{
		Data:     videodata,
		Duration: time.Second / 30,
	}); videoErr != nil {
		log.Debug("WriteSample err", videoErr)
		return fmt.Errorf("WriteSample err %s", videoErr)
	}
	return nil
}

func (p *Peer) SendPeerRTPAudio(audiodata []byte) error {
	if p.status != PEER_CONNECT {
		// return fmt.Errorf("peer is not connected,status %d", p.status)
		// log.Debug("SendPeerAudio peer ", p.peerName, " is not connected,status ", p.status)
		return nil
	}
	if p.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
		p.status = PEER_CLOSED
		p.endTime = time.Now()
		log.Debug("SendPeerAudio peer ", p.peerName, " is closed,status ", p.status)
		return nil
	}
	if config.Config.Stream.Debug {
		log.Debug("SendPeerAudio peer name:", p.peerName, "stream name:", p.streamName, "opus len:", len(audiodata))
	}
	if p.audioRTPTrack == nil {
		return fmt.Errorf("peer audioRTPTrack is nil")
	}
	if _, audioErr := p.audioRTPTrack.Write(audiodata); audioErr != nil {
		log.Debug("WriteSample err", audioErr)
		return fmt.Errorf("WriteSample err %s", audioErr)
	}

	return nil
}
func (p *Peer) SendPeerRtpVideo(videodata []byte) error {
	if p.status != PEER_CONNECT {
		// return fmt.Errorf("peer is not connected,status %d", p.status)
		// log.Debug("SendPeerVideo peer ", p.peerName, " is not connected,status ", p.status)
		return nil
	}
	if p.peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
		p.status = PEER_CLOSED
		p.endTime = time.Now()
		// return fmt.Errorf("peer is closed,status %d", p.status)
		log.Debug("SendPeerVideo peer ", p.peerName, " is closed,status ", p.status)
		return nil
	}
	if p.videoRTPTrack == nil {
		return fmt.Errorf("peer videoRTPTrack is nil")
	}
	if _, videoErr := p.videoRTPTrack.Write(videodata); videoErr != nil {
		log.Debug("WriteSample err", videoErr)
		return fmt.Errorf("WriteSample err %s", videoErr)
	}
	return nil
}

type streamPeerinterface interface {
	InitStream(streamid string,
		streamname string,
		username string,
		password string)
	AddPeer(*Peer) error
	DeletePeer(Peer) error
	GetPeer(string) (*Peer, error)
	SetPeer(string, *Peer) error
	SetOpusCtl()
	InitAudio([]byte) error
	InitVideo([]byte) error
	SendStreamAudio([]byte) error
	SendStreamVideo([]byte) error
	// SetAudiodecoder(*fdkaac.AacDecoder) (error)
}

//rtmp 流
type Stream struct {
	streamId       string
	streamName     string
	streamType     string
	userName       string
	passWord       string
	status         StreamState
	startTime      time.Time
	endTime        time.Time
	peers          map[string]*Peer
	audioDecoder   *fdkaac.AacDecoder
	audioEncoder   *opus.Encoder
	remotetrack    []*webrtc.TrackRemote
	audioBuffer    []byte
	audioClockRate uint32
	//room                   *lksdk.Room //for publish to livekit room ,first create room as streamname for livekit,then publish track to livekit
	room *livekitclient.Room
	// videoTrack, audioTrack *webrtc.TrackLocalStaticSample
	streamPeerinterface
}

func (s *Stream) SetRemoteTrack(remotetrack *webrtc.TrackRemote) {
	s.remotetrack = append(s.remotetrack, remotetrack)
}
func (s *Stream) GetRemoteTrack() ([]*webrtc.TrackRemote, error) {
	if s.remotetrack == nil {
		return nil, errors.New(" not remotetrack")
	} else {
		return s.remotetrack, nil
	}
}
func (s *Stream) InitStream(streamid string,
	streamname string,
	username string,
	password string,
	streamtype string) {
	s.streamId = streamid
	s.streamName = streamname
	s.userName = username
	s.passWord = password
	s.streamType = streamtype
	s.peers = make(map[string]*Peer)
	s.status = STREAM_INIT
	s.startTime = time.Now()
}
func (s *Stream) AddPeer(p *Peer) error {
	if s.peers == nil || p == nil {
		s.peers = make(map[string]*Peer, 0)
	}
	if s.peers[p.peerName] != nil {
		log.Debug(p.peerName, " peer is exsit,reset peer")
	}
	if p.peerName != "" {
		// p.status = PEER_CONNECT
		// p.startTime = time.Now()
		s.peers[p.peerName] = p
		return nil
	}
	return errors.New("peer is null")
}
func (s *Stream) DeletePeer(p *Peer) error {
	if s.peers == nil || s.peers[p.peerName] == nil || p == nil {
		return errors.New("peers is not exsit")
	}
	if p.peerName != "" {
		delete(s.peers, p.peerName)
		return nil
	}

	return errors.New("peer is not exsit")
}

func (s *Stream) GetPeer(peername string) (*Peer, error) {
	if s.peers == nil {
		return nil, errors.New("peers is not exsit")
	}
	if p := s.peers[peername]; p == nil {
		return nil, fmt.Errorf("peer %s is not exsit", peername)
	} else {
		return p, nil
	}
}
func (s *Stream) SetPeer(peername string, pp *Peer) error {
	if s.peers == nil || pp == nil {
		return errors.New("peers is not exsit")
	}
	if p := s.peers[peername]; p == nil || pp.peerName != peername {
		return fmt.Errorf("peer %s is not exsit", peername)
	} else {
		s.peers[peername] = pp
		return nil
	}
}

func (s *Stream) SetOpusCtl() {
	s.audioEncoder.SetMaxBandwidth(opus.Bandwidth(2))
	s.audioEncoder.SetComplexity(9)
	s.audioEncoder.SetBitrateToAuto()
	s.audioEncoder.SetInBandFEC(true)
}
func (s *Stream) InitAudio(data []byte) error {
	encoder, err := opus.NewEncoder(48000, 2, opus.AppAudio)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	s.audioEncoder = encoder
	s.SetOpusCtl()
	s.audioDecoder = fdkaac.NewAacDecoder()
	s.audioDecoder.InitRaw(data)
	return nil
}
func (s *Stream) ReleaseAudio() {
	s.audioDecoder.Close()
	s.audioEncoder.Close()
}
func (s *Stream) SendStreamAudio(datas []byte) []error {
	var errs []error
	if s.audioDecoder == nil {
		log.Debug("decoder is released")
		errs = append(errs, fmt.Errorf("decoder is released"))
		return errs
	}
	pcm, err := s.audioDecoder.Decode(datas)
	if err != nil {
		log.Debug("decode error: ", hex.EncodeToString(datas), err)
		errs = append(errs, fmt.Errorf("decode error"))
		return errs
	}
	if config.Config.Stream.Debug {
		log.Debug("\r\npcm len ", len(pcm), " ->") //, pcm)
	}
	blockSize := 960
	for s.audioBuffer = append(s.audioBuffer, pcm...); len(s.audioBuffer) >= blockSize*4; s.audioBuffer = s.audioBuffer[blockSize*4:] {
		pcm16 := make([]int16, blockSize*2)
		pcm16len := len(pcm16)
		for i := 0; i < pcm16len; i++ {
			pcm16[i] = int16(binary.LittleEndian.Uint16(s.audioBuffer[i*2:]))
		}
		bufferSize := 1024
		opusData := make([]byte, bufferSize)
		if s.audioEncoder == nil {
			log.Debug("encoder is released")
			errs = append(errs, fmt.Errorf("encoder is released"))
			return errs
		}
		n, err := s.audioEncoder.Encode(pcm16, opusData)
		// n, err := h.audioEncoder.ReadEncode(pcm16, opusData)
		if err != nil {
			errs = append(errs, err)
			return errs
		}
		opusOutput := opusData[:n]
		// m:=GetRoom
		// room, err := h.streammanager.GetRoom("")
		if s.room == nil {
			//log.Debug("stream room is null")
		} else {
			//960 48k 20ms/per
			if s.streamType != "RTP" {
				//send to sfu server
				if config.Config.Stream.IONSfuEnable {
					s.room.TrackSendIonData(s.streamName, "audio", opusOutput, 20*time.Millisecond)
				}
				if config.Config.Stream.LiveKitSfuEnable {
					s.room.TrackSendLivekitData(s.streamName, "audio", opusOutput, 20*time.Millisecond)
				}
			}
		}
		for pname, p := range s.peers {
			if config.Config.Stream.Debug {
				log.Debug("peer ", pname)
			}
			if p.streamName == s.streamName {
				//log.Printf(" send audio data ")

				err := p.SendPeerAudio(opusOutput)
				if err != nil {
					log.Debug("error", err)
					errs = append(errs, err)
				}

			}
		}

	}
	return errs
}

func (s *Stream) SendStreamAudioFromWebrtc(datas []byte) []error {
	var errs []error

	opusOutput := datas
	// m:=GetRoom
	// room, err := h.streammanager.GetRoom("")
	if s.room != nil && s.status == STREAM_CONNECT {
		//log.Debug("stream room is null")
		// } else {
		//960 48k 20ms/per
		if s.streamType == "RTP" {

			//send audio stream to ion sfu server
			if config.Config.Stream.IONSfuEnable {
				s.room.TrackSendIonRtpPackets(s.streamName, "audio", opusOutput)
			}
			//send audio stream to livekit sfu server
			if config.Config.Stream.LiveKitSfuEnable {
				s.room.TrackSendLivekitRtpPackets(s.streamName, "audio", opusOutput)
			}
		} else {
			if config.Config.Stream.IONSfuEnable {
				s.room.TrackSendIonData(s.streamName, "audio", opusOutput, 20*time.Millisecond)
			}
			if config.Config.Stream.LiveKitSfuEnable {
				s.room.TrackSendLivekitData(s.streamName, "audio", opusOutput, 20*time.Millisecond)
			}
		}
	}
	for pname, p := range s.peers {
		if config.Config.Stream.Debug {
			log.Debug("peer ", pname)
		}
		if p.streamName == s.streamName {
			//log.Printf(" send audio data ")
			if s.streamType == "RTP" {
				err := p.SendPeerRTPAudio(opusOutput)
				if err != nil {
					log.Debug("error", err)
					errs = append(errs, err)
				}
			} else {
				err := p.SendPeerAudio(opusOutput)
				if err != nil {
					log.Debug("error", err)
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}

func (s *Stream) IsRtpStream() bool {
	return s.streamType == "RTP"
}
func (s *Stream) SendStreamVideo(datas []byte) []error {
	var errs []error
	if s.streamType == "RTP" {
		if s.room != nil && s.status == STREAM_CONNECT {
			//send rtp stream to ion sfu
			if config.Config.Stream.IONSfuEnable {
				s.room.TrackSendIonRtpPackets(s.streamName, "video", datas)
			}
			//send rtp stream to livekit sfu
			if config.Config.Stream.LiveKitSfuEnable {
				s.room.TrackSendLivekitRtpPackets(s.streamName, "video", datas)
			}
		}
	} else {
		if s.room != nil && s.status == STREAM_CONNECT {
			if config.Config.Stream.IONSfuEnable {
				s.room.TrackSendIonData(s.streamName, "video", datas, time.Second/30)
			}
			if config.Config.Stream.LiveKitSfuEnable {
				s.room.TrackSendLivekitData(s.streamName, "video", datas, time.Second/30)
			}
		}
	}
	for pname, p := range s.peers {
		if config.Config.Stream.Debug {
			log.Debug("peer ", pname)
		}
		if p.streamName == s.streamName {
			if config.Config.Stream.Debug {
				log.Debug(" send video data ")
			}
			if s.streamType == "RTP" {
				err := p.SendPeerRtpVideo(datas)
				if err != nil {
					log.Debug("error", err)
					errs = append(errs, err)
				}
			} else {
				err := p.SendPeerVideo(datas)
				if err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	return errs
}

type streamsinterface interface {
	InitStreamManage(ctx context.Context)
	AddStream(*Stream) error
	DeleteStream(*Stream) error
	GetStream(string) (*Stream, error)
	SetStream(string, *Stream) error
}

//流管理
type StreamManager struct {
	streams    map[string]*Stream
	roomMap    map[string]*livekitclient.Room //多房间管理
	deviceroom *livekitclient.Device_Room     //设备房间对应关系
	streamsinterface
	ctx context.Context
	// livekitclient.Room
}

func (m *StreamManager) InitStreamManage(ctx context.Context) {
	m.streams = make(map[string]*Stream, 0)
	m.roomMap = make(map[string]*livekitclient.Room)
	m.ctx = ctx

	room := livekitclient.NewRoom(ctx, &config.Config.Livekit.Token)
	// sn, _ := identity.GetSN()
	//create livekit room with this server's Identity like mac addr
	if config.Config.Stream.LiveKitSfuEnable {
		roomname := config.Config.Livekit.Token.Identity // + "->LIVEKIT"
		_, err := room.CreateliveKitRoom(roomname)
		if err != nil {
			log.Debug("livekit room create failed: ", err)
		}
		roomname = config.Config.Livekit.Token.Identity
		m.roomMap[roomname] = room
	}
	//create ion room  with this server's Identity like mac addr
	if config.Config.Stream.IONSfuEnable {
		roomname := config.Config.Livekit.Token.Identity // + "->ION"
		_, err := room.CreateIonRoom(config.Config.Livekit.Token.HostIon, roomname)
		if err != nil {
			log.Debug("ion room create failed: ", err)
		}
		roomname = config.Config.Livekit.Token.Identity
		m.roomMap[roomname] = room
	}

	m.deviceroom = livekitclient.NewDevice_Room()

}
func (m *StreamManager) GetDefaultRoom() *livekitclient.Room {

	return m.roomMap[config.Config.Livekit.Token.Identity]
}
func (m *StreamManager) GetRoomByName(roomname string) *livekitclient.Room {
	return m.roomMap[roomname]
}
func (m *StreamManager) AddStream(s *Stream) error {
	if m.streams == nil || s == nil {
		return errors.New("stream is not exsit")
	}
	if s.streamName != "" {
		// m.streams[s.streamName] = s
		//原来存在这个源就直接改状态
		if ss := m.streams[s.streamName]; ss == nil {
			s.room = m.GetDefaultRoom()
			s.status = STREAM_CONNECT
			m.streams[s.streamName] = s
		} else {
			ss.room = m.GetDefaultRoom()
			ss.startTime = time.Now()
			ss.status = STREAM_CONNECT
			m.streams[s.streamName] = ss
		}
		if s.streamType == "RTP" {
			//publish to ion sfu
			if config.Config.Stream.IONSfuEnable {
				m.GetDefaultRoom().RTPTrackPublished_to_ION(s.remotetrack, s.streamName)
			}
			//publish to livekit sfu
			if config.Config.Stream.LiveKitSfuEnable {
				m.GetDefaultRoom().RTPTrackPublished(s.remotetrack, s.streamName)
			}
		} else {
			if config.Config.Stream.IONSfuEnable {
				m.GetDefaultRoom().TrackPublished_to_ION(s.streamName)
			}
			if config.Config.Stream.LiveKitSfuEnable {
				m.GetDefaultRoom().TrackPublished(s.streamName)
			}
		}
		return nil
	} else {
		return errors.New("stream is not exsit")
	}

}
func (m *StreamManager) DeleteStream(name string) error {
	if m.streams == nil || name == "" {
		return errors.New("stream is not exsit")
	}
	s := m.streams[name]

	if s != nil {
		s.status = STREAM_DEADLINE
		if config.Config.Stream.LiveKitSfuEnable {
			m.GetDefaultRoom().TrackClose(s.streamName)
		}
		if config.Config.Stream.IONSfuEnable {
			m.GetDefaultRoom().TrackCloseION(s.streamName)
		}
		// go func() { //防止正在使用，先设置状态，然后延迟再删除，如果还有冲突，就先不删除，一般是在stream推流关闭时才调用一般不会出问题
		// 	time.Sleep(time.Duration(2) * time.Second)
		// 	s.ReleaseAudio()
		// 	delete(m.streams, name)
		// }()
		return nil
	} else {
		return errors.New("stream is not exsit")
	}

}
func (m *StreamManager) GetStream(streamname string) (*Stream, error) {
	if m.streams == nil {
		return nil, errors.New("stream is not exsit")
	}
	if streamname != "" {
		return m.streams[streamname], nil
	} else {
		return nil, errors.New("streamname is null")
	}
}
func (m *StreamManager) GetRoom(roomname string) (*livekitclient.Room, error) {
	if m.roomMap[roomname] == nil {
		return nil, fmt.Errorf("room(%s) is not exsit", roomname)
	}
	return m.roomMap[roomname], nil
}
func (m *StreamManager) GetDeviceRoom() *livekitclient.Device_Room {
	return m.deviceroom
}
func (m *StreamManager) SetStream(name string, s *Stream) error {
	if m.streams == nil || s == nil {
		return errors.New("stream is not exsit")
	}
	if name != "" && name == s.streamName {
		m.streams[name] = s
		return nil
	} else {
		return errors.New("streamname is null")
	}
}
func (m *StreamManager) EndRTMP() {
	room := m.GetDefaultRoom()
	if room != nil {
		room.Close()
	}
}
func (m *StreamManager) EndAll() {
	for _, room := range m.roomMap {
		room.Close()
	}
}

//全局变量
var (
	Global_StreamM        StreamManager
	StreamPeersForConnect map[string]*Peer //不存在的流，考虑收到peer时通过mqtt命令向设备发起推流指令，比较好的策略是有客户端向设备询问是否在推流
)

func CreateGlobalStreamM(ctx context.Context) *StreamManager {
	Global_StreamM.InitStreamManage(ctx)
	return &Global_StreamM
}
func GetGlobalStreamM() *StreamManager {

	return &Global_StreamM
}
