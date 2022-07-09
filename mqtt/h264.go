package mqtt

import (
	"bytes"
	"encoding/binary"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pkg/errors"
	"github.com/xiangxud/rtmp_webrtc_server/log"
)

type H264Packet struct {
	caehedPacket *codecs.H264Packet
	hasKeyFrame  bool
}

func UnpackRTP2H264(rtpPayload []byte) ([]byte, bool) {

	if len(rtpPayload) <= 0 {

		return nil, false

	}

	var out []byte

	fu_indicator := rtpPayload[0] //获取第一个字节

	fu_header := rtpPayload[1] //获取第二个字节

	nalu_type := fu_indicator & 0x1f //获取FU indicator的类型域

	flag := fu_header & 0xe0 //获取FU header的前三位，判断当前是分包的开始、中间或结束

	nal_fua := ((fu_indicator & 0xe0) | (fu_header & 0x1f)) //FU_A nal

	var FrameType string

	if nal_fua == 0x67 {

		FrameType = "SPS"

	} else if nal_fua == 0x68 {

		FrameType = "PPS"

	} else if nal_fua == 0x65 {

		FrameType = "IDR"

	} else if nal_fua == 0x61 {

		FrameType = "P Frame"

	} else if nal_fua == 0x41 {

		FrameType = "P Frame"

	}
	// lens := len(rtpPayload)
	// log.Debug("payload len ", lens)
	// if lens > 10 {
	// 	lens = 10
	// }

	// log.InfoHex(rtpPayload, lens)
	log.Debug("nalu_type: ", nalu_type, " flag: ", flag, " FrameType: ", FrameType)

	if nalu_type == 0x1c { //判断NAL的类型为0x1c=28，说明是FU-A分片

		if flag == 0x80 { //分片NAL单元开始位

			/*

			   o := make([]byte, len(rtpPayload)+5-2) //I帧开头可能为00 00 00 01、00 00 01，组帧时只用00 00 01开头

			   o[0] = 0x00

			   o[1] = 0x00

			   o[2] = 0x00

			   o[3] = 0x01

			   o[4] = nal_fua*/

			o := make([]byte, len(rtpPayload)+4-2) //I帧开头可能为00 00 00 01、00 00 01，组帧时只用00 00 01开头

			o[0] = 0x00

			o[1] = 0x00

			o[2] = 0x01

			o[3] = nal_fua

			copy(o[4:], rtpPayload[2:])

			out = o

			return out, true

		} else { //中间分片包或者最后一个分片包

			o := make([]byte, len(rtpPayload)-2)

			copy(o[0:], rtpPayload[2:])

			out = o
			if len(out) < 1394 {
				return out, false
			} else {
				return out, true
			}

		}

	} else if nalu_type == 0x1 { //单一NAL 单元模式

		o := make([]byte, len(rtpPayload)+4) //将整个rtpPayload一起放进去

		o[0] = 0x00

		o[1] = 0x00

		o[2] = 0x00

		o[3] = 0x01

		copy(o[4:], rtpPayload[0:])

		out = o

	} else {

		log.Debug("Unsport nalu type!")

	}

	return out, false

}
func (h *H264Packet) GetRTPRawH264(packet *rtp.Packet) ([]byte, error) {
	// nalPrefix3Bytes := []byte{0, 0, 1}
	// nalPrefix4Bytes := []byte{0, 0, 0, 1}
	//  var p_len uint32
	p_len := len(packet.Payload)
	if p_len == 0 {
		return nil, errors.New("zero packet payload")
	}
	// if !h.hasKeyFrame {
	// 	if h.hasKeyFrame = isKeyFrame(packet.Payload); !h.hasKeyFrame {
	// 		return nil, errors.New("not key frame")
	// 	}
	// }
	if h.caehedPacket == nil {
		h.caehedPacket = &codecs.H264Packet{}
	}
	// if len(packet.Payload) >= 1200 {
	// 	log.Debug("packet is extra payload")
	// }

	if p_len < 5 {
		log.Debug(packet.Payload[:p_len])
	} else {
		log.Debug(packet.Payload[:5])
	}
	data, err := h.caehedPacket.Unmarshal(packet.Payload)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func isKeyFrame(data []byte) bool {
	const (
		typeSTAPA       = 24
		typeSPS         = 7
		naluTypeBitmask = 0x1F
	)

	var word uint32

	payload := bytes.NewReader(data)
	if err := binary.Read(payload, binary.BigEndian, &word); err != nil {
		return false
	}

	naluType := (word >> 24) & naluTypeBitmask
	if naluType == typeSTAPA && word&naluTypeBitmask == typeSPS {
		return true
	} else if naluType == typeSPS {
		return true
	}

	return false
}
