package media

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Glimesh/go-fdkaac/fdkaac"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	opus "github.com/xiangxud/rtmp_webrtc_server/opus"
)

type StreamState byte

const (
	PEER_INIT StreamState = iota
	PEER_CONNECT
	PEER_TIMEOUT
	PEER_DEADLINE
)
const (
	STREAM_INIT StreamState = iota
	STREAM_CONNECT
	STREAM_TIMEOUT
	STREAM_DEADLINE
)

type peerInterface interface {
	InitPeer(peerid int64,
		peername string,
		username string,
		password string) error
	AddConnect(*webrtc.PeerConnection)
	AddAudioTrack(*webrtc.TrackLocalStaticSample)
	AddVideoTrack(*webrtc.TrackLocalStaticSample)
	SendPeerAudio([]byte) error
	SendPeerVideo([]byte) error
	VerifyPeer()
}

//webrtc 客户
type Peer struct {
	peerId                 string
	peerName               string
	userName               string
	streamName             string
	passWord               string
	status                 StreamState
	startTime              time.Time
	endTime                time.Time
	peerConnection         *webrtc.PeerConnection
	videoTrack, audioTrack *webrtc.TrackLocalStaticSample
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

}
func (p *Peer) AddAudioTrack(track *webrtc.TrackLocalStaticSample) {
	p.audioTrack = track
}
func (p *Peer) AddVideoTrack(track *webrtc.TrackLocalStaticSample) {
	p.videoTrack = track
}
func (p *Peer) SendPeerAudio(audiodata []byte) error {
	log.Println("SendPeerAudio peer name:", p.peerName, "stream name:", p.streamName, "opus len:", len(audiodata))
	if audioErr := p.audioTrack.WriteSample(media.Sample{
		Data:     audiodata,
		Duration: 20 * time.Millisecond,
	}); audioErr != nil {
		log.Println("WriteSample err", audioErr)
		return fmt.Errorf("WriteSample err %s", audioErr)
	}
	return nil
}
func (p *Peer) SendPeerVideo(videodata []byte) error {
	if vedioErr := p.videoTrack.WriteSample(media.Sample{
		Data:     videodata,
		Duration: time.Second / 30,
	}); vedioErr != nil {
		log.Println("WriteSample err", vedioErr)
		return fmt.Errorf("WriteSample err %s", vedioErr)
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
	userName       string
	passWord       string
	status         StreamState
	startTime      time.Time
	endTime        time.Time
	peers          map[string]*Peer
	audioDecoder   *fdkaac.AacDecoder
	audioEncoder   *opus.Encoder
	audioBuffer    []byte
	audioClockRate uint32
	streamPeerinterface
}

func (s *Stream) InitStream(streamid string,
	streamname string,
	username string,
	password string) {
	s.streamId = streamid
	s.streamName = streamname
	s.userName = username
	s.passWord = password
	s.peers = make(map[string]*Peer)
	s.status = STREAM_INIT
	s.startTime = time.Now()
}
func (s *Stream) AddPeer(p *Peer) error {
	if s.peers == nil || p == nil {
		s.peers = make(map[string]*Peer, 0)
	}
	if s.peers[p.peerName] != nil {
		return errors.New("peer is exsit")
	}
	if p.peerName != "" {
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
		log.Println(err.Error())
		return err
	}
	s.audioEncoder = encoder
	s.SetOpusCtl()
	s.audioDecoder = fdkaac.NewAacDecoder()
	s.audioDecoder.InitRaw(data)
	return nil
}
func (s *Stream) SendStreamAudio(datas []byte) []error {
	var errs []error
	pcm, err := s.audioDecoder.Decode(datas)
	if err != nil {
		log.Println("decode error: ", hex.EncodeToString(datas), err)
		errs = append(errs, fmt.Errorf("decode error"))
		return errs
	}
	log.Println("\r\npcm len ", len(pcm), " ->") //, pcm)
	blockSize := 960
	for s.audioBuffer = append(s.audioBuffer, pcm...); len(s.audioBuffer) >= blockSize*4; s.audioBuffer = s.audioBuffer[blockSize*4:] {
		pcm16 := make([]int16, blockSize*2)
		pcm16len := len(pcm16)
		for i := 0; i < pcm16len; i++ {
			pcm16[i] = int16(binary.LittleEndian.Uint16(s.audioBuffer[i*2:]))
		}
		bufferSize := 1024
		opusData := make([]byte, bufferSize)
		n, err := s.audioEncoder.Encode(pcm16, opusData)
		// n, err := h.audioEncoder.ReadEncode(pcm16, opusData)
		if err != nil {
			errs = append(errs, err)
			return errs
		}
		opusOutput := opusData[:n]
		for pname, p := range s.peers {
			log.Println("peer ", pname)
			if p.streamName == s.streamName {
				//log.Printf(" send audio data ")
				err := p.SendPeerAudio(opusOutput)
				if err != nil {
					log.Println("error", err)
					errs = append(errs, err)
				}
			}
		}

	}
	return errs
}

func (s *Stream) SendStreamVideo(datas []byte) []error {
	var errs []error
	for pname, p := range s.peers {
		log.Println("peer ", pname)
		if p.streamName == s.streamName {
			log.Printf(" send video data ")
			err := p.SendPeerVideo(datas)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

type streamsinterface interface {
	InitStreamManage()
	AddStream(*Stream) error
	DeleteStream(*Stream) error
	GetStream(string) (*Stream, error)
	SetStream(string, *Stream) error
}

//流管理
type StreamManager struct {
	streams map[string]*Stream
	streamsinterface
}

func (m *StreamManager) InitStreamManage() {
	m.streams = make(map[string]*Stream, 0)
}
func (m *StreamManager) AddStream(s *Stream) error {
	if m.streams == nil || s == nil {
		return errors.New("stream is not exsit")
	}
	if s.streamName != "" {
		m.streams[s.streamName] = s
		return nil
	} else {
		return errors.New("stream is not exsit")
	}

}
func (m *StreamManager) DeleteStream(s *Stream) error {
	if m.streams == nil || s == nil {
		return errors.New("stream is not exsit")
	}
	if s.streamName != "" {
		delete(m.streams, s.streamName)
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

var Global_StreamM StreamManager

func CreateGlobalStreamM() *StreamManager {
	Global_StreamM.InitStreamManage()
	return &Global_StreamM
}
func GetGlobalStreamM() *StreamManager {

	return &Global_StreamM
}
