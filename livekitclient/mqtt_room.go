//作为mqtt信令和grpc信令的桥梁
package livekitclient

import (
	"errors"

	"github.com/pion/webrtc/v3"
	// "github.com/xiangxud/rtmp_webrtc_server/mqtt"
)

type Device struct {
	Deviceid string `json:"deviceid"`
	// Devicename string `json:"devicename"`
	Streamname     string `json:"streamname"`
	PRoomList      map[string]*Room
	RemoteSDPOffer webrtc.SessionDescription `json:"sdpoffer"`
	LocalSDPanswer webrtc.SessionDescription `json:"answer"`
	// VideoTrack     *webrtc.TrackLocalStaticSample
	// AudioTrack     *webrtc.TrackLocalStaticSample
}

type Device_Room struct {
	// PRoom       map[string]*Room //房间统一由服务器管理，由客户端指定房间token，默认为本服务器自身sn创建的房间
	DevicesList map[string]*Device
}

// var (
// 	g_DeviceRoom Device_Room
// )

// func init() {
// 	// g_DeviceRoom.PRoom = make(map[string]*Room, 0)
// 	g_DeviceRoom.DevicesList = make(map[string]*Device, 0)
// }
// func GetGlobalDeviceRoom() *Device_Room {
// 	return &g_DeviceRoom
// }
func NewDevice(deviceid, devicename, streamname string) *Device {
	pd := &Device{
		Deviceid: deviceid,
		// Devicename: devicename,
		Streamname: streamname,
	}
	pd.PRoomList = make(map[string]*Room)
	return pd
}
func (d *Device) SetStreamname(streamname string) {
	d.Streamname = streamname
}
func NewDevice_Room() *Device_Room {
	dv := Device_Room{}
	// dv.PRoom = make(map[string]*Room, 0)
	dv.DevicesList = make(map[string]*Device, 0)
	return &dv
}

// func (dr *Device_Room) AddRoom(proom *Room) error {
// 	if proom != nil && dr.PRoom[proom.Identity] != nil {
// 		dr.PRoom[proom.Identity] = proom
// 		return nil
// 	} else {
// 		return errors.New("room is nil")
// 	}
// }
// func (dr *Device_Room) DelRoom(proom *Room) error {
// 	if proom != nil && dr.PRoom[proom.Identity] != nil {
// 		dr.PRoom[proom.Identity] = proom
// 		return nil
// 	} else {
// 		return errors.New("room is nil")
// 	}
// }
func (dr *Device_Room) AddDevice(pdevice *Device) error {
	if pdevice != nil && dr.DevicesList[pdevice.Deviceid] != nil {
		dr.DevicesList[pdevice.Deviceid] = pdevice
		return nil
	} else {
		return errors.New("device is nil")
	}
}
func (dr *Device_Room) DelDevice(deviceid string) error {
	if deviceid != "" && dr.DevicesList[deviceid] != nil {
		dr.DevicesList[deviceid] = nil
		return nil
	} else {
		return errors.New("device is not exsit")
	}
}
func (dr *Device_Room) GetDevice(deviceid string) (*Device, error) {
	if deviceid != "" {
		if d := dr.DevicesList[deviceid]; d != nil {

			return d, nil
		} else {
			return nil, errors.New("device is not exsit")
		}
	}
	return nil, errors.New("device is not exsit")
}

func (d *Device) SetOffer(offer webrtc.SessionDescription) {
	d.RemoteSDPOffer = offer
}
func (d *Device) SetAnswer(answer webrtc.SessionDescription) {
	d.LocalSDPanswer = answer
}

//断开某设备的房间号
func (dr *Device_Room) EndDeviceByDeviceName(devicename string) error {
	device, err := dr.GetDevice(devicename)
	if device == nil {
		return err
	}
	for _, proom := range device.PRoomList {
		if proom != nil {
			proom.Close()
		}
	}
	return nil
}

//断开某一个房间号
func (dr *Device_Room) EndDeviceByRoomName(roomname string) error {
	for _, device := range dr.DevicesList {
		if device == nil {
			continue
		}
		for _, proom := range device.PRoomList {
			if proom != nil {
				if proom.RoomName == roomname {
					proom.Close()
				}
			}
		}
	}
	return nil
}

//断开所有设备的房间号
func (dr *Device_Room) EndAll() {
	for _, device := range dr.DevicesList {
		if device == nil {
			continue
		}
		for _, proom := range device.PRoomList {
			if proom != nil {
				proom.Close()
			}
		}
	}
}
