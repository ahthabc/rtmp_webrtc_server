package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
	// "github.com/xiangxud/go-rtmp"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"

	// "github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
	// rtmpmsg "github.com/xiangxud/go-rtmp/message"
	// opus "gopkg.in/hraban/opus.v2"
)

func startRTMPServer(streammanager *media_interface.StreamManager) {
	log.Println("Starting RTMP Server")

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1935")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			return conn, &rtmp.ConnConfig{
				Handler: &Handler{
					streammanager: streammanager,
					streamname:    "",
				},

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},
			}
		},
	})
	if err := srv.Serve(listener); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}

type Handler struct {
	rtmp.DefaultHandler
	streammanager *media_interface.StreamManager
	streamname    string
}

func (h *Handler) OnServe(conn *rtmp.Conn) {

}

func (h *Handler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Printf("OnConnect: %#v", cmd)
	// h.audioClockRate = 48000
	return nil
}

func (h *Handler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Printf("OnCreateStream: %#v", cmd)
	return nil
}

func (h *Handler) OnPublish(timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Printf("OnPublish: %#v", cmd)

	if cmd.PublishingName == "" {
		return errors.New("PublishingName is empty")
	}
	s := media_interface.Stream{}
	s.InitStream(cmd.PublishingType, cmd.PublishingName, "", "")
	// s.InitAudio()
	// s.InitVideo()
	h.streamname = cmd.PublishingName
	log.Println("current streamname:", h.streamname)
	m := media_interface.GetGlobalStreamM()
	err := m.AddStream(&s)
	if err != nil {
		log.Println("addstream error", err)
		return err
	}
	return nil
}

func (h *Handler) OnAudio(timestamp uint32, payload io.Reader) error {
	// Convert AAC to opus
	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		return err
	}
	if data.Len() <= 0 {
		log.Println("no audio datas", timestamp, payload)
		return fmt.Errorf("no audio datas")
	}
	// log.Println("\r\ntimestamp->", timestamp, "\r\npayload->", payload, "\r\naudio data->", data.Bytes())
	datas := data.Bytes()
	// log.Println("\r\naudio data len:", len(datas), "->") // hex.EncodeToString(datas))

	stream, err := h.streammanager.GetStream(h.streamname)
	if err != nil {
		log.Println(err, "error Get current Stream ")
		return fmt.Errorf("can't initialize codec with %s", err.Error())
	}
	if audio.AACPacketType == flvtag.AACPacketTypeSequenceHeader {
		log.Println("Created new codec ", hex.EncodeToString(datas))

		err := stream.InitAudio(datas)
		if err != nil {
			log.Println(err, "error initializing Audio")
			return fmt.Errorf("can't initialize codec with %s", err.Error())
		}
		// err = stream.audioDecoder.InitRaw(datas)

		if err != nil {
			log.Println(err, "error initializing stream")
			return fmt.Errorf("can't initialize codec with %s", hex.EncodeToString(datas))
		}

		return nil
	}

	errs := stream.SendStreamAudio(datas)

	if errs != nil {
		var errstr string
		for _, e := range errs {
			errstr = errstr + e.Error()
		}
		return fmt.Errorf("send audio error: %s", errstr)
	}
	return nil

	// return nil
}
func (h *Handler) OnAudioPCMA(timestamp uint32, payload io.Reader) error {
	var audio flvtag.AudioData
	if err := flvtag.DecodeAudioData(payload, &audio); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, audio.Data); err != nil {
		return err
	}
	return nil
	// return h.audioTrack.WriteSample(media.Sample{
	// 	Data:     data.Bytes(),
	// 	Duration: 128 * time.Millisecond,
	// })
}

const headerLengthField = 4

func (h *Handler) OnVideo(timestamp uint32, payload io.Reader) error {
	var video flvtag.VideoData
	if err := flvtag.DecodeVideoData(payload, &video); err != nil {
		return err
	}

	data := new(bytes.Buffer)
	if _, err := io.Copy(data, video.Data); err != nil {
		return err
	}

	outBuf := []byte{}
	videoBuffer := data.Bytes()
	for offset := 0; offset < len(videoBuffer); {
		bufferLength := int(binary.BigEndian.Uint32(videoBuffer[offset : offset+headerLengthField]))
		if offset+bufferLength >= len(videoBuffer) {
			break
		}

		offset += headerLengthField
		outBuf = append(outBuf, []byte{0x00, 0x00, 0x00, 0x01}...)
		outBuf = append(outBuf, videoBuffer[offset:offset+bufferLength]...)

		offset += int(bufferLength)
	}
	stream, err := h.streammanager.GetStream(h.streamname)
	if err != nil {
		log.Println(err, "error Get current Stream ")
		return fmt.Errorf("can't initialize codec with %s", err.Error())
	}
	errs := stream.SendStreamVideo(outBuf)
	if errs != nil {
		var errstr string
		for _, e := range errs {
			errstr = errstr + fmt.Sprintf("%s-", e.Error())
		}
		return fmt.Errorf("send video error: %s", errstr)
	}
	return nil
	// return h.videoTrack.WriteSample(media.Sample{
	// 	Data:     outBuf,
	// 	Duration: time.Second / 30,
	// })
}

func (h *Handler) OnClose() {
	log.Printf("OnClose")
	m := media_interface.GetGlobalStreamM()
	err := m.DeleteStream(h.streamname)
	if err != nil {
		log.Println("addstream error", err)
		//return err
	}
}
