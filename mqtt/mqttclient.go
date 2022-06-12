package mqtt

// Connect to the broker, subscribe, and write messages received to a file
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	//"github.com/didi/nightingale/src/modules/agent/config"
	//"github.com/didi/nightingale/src/modules/agent/wol"
	// "github.com/didi/nightingale/src/modules/agent/config"
	// "github.com/didi/nightingale/src/modules/agent/kvm/utils"
	// "github.com/didi/nightingale/src/modules/agent/report"
	"github.com/xiangxud/rtmp_webrtc_server/config"
	"github.com/xiangxud/rtmp_webrtc_server/identity"
	"github.com/xiangxud/rtmp_webrtc_server/log"
	media_interface "github.com/xiangxud/rtmp_webrtc_server/media"
	enc "github.com/xiangxud/rtmp_webrtc_server/signal"

	// "honnef.co/go/tools/config"
	// "github.com/didi/nightingale/src/modules/agent/wol"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pion/webrtc/v3"
)

const (
	CMDMSG_OFFER      = "offer"
	CMDMSG_ANSWER     = "answer"
	CMDMSG_ERROR      = "error"
	CMDMSG_STOPPEER   = "stoppeer"
	CMDMSG_RESUMEPEER = "resumepeer"
	CMDMSG_DELETEPEER = "deletepeer"
	MODE_RTMP         = "rtmp"
	TOPIC_ANSWER      = "answer"
	TOPIC_ERROR       = "error"
	TOPIC_REFUSE      = "refuse"
)

var (
	msgChans chan PublishMsg //prompb.WriteRequest //multi node one chan
)

type Session struct {
	Type     string `json:"type"`
	Msg      string `json:"msg"`
	Data     string `json:"data"`
	DeviceId string `json:"device_id"`
}
type Message struct {
	SeqID              string                    `json:"seqid"`
	Mode               string                    `json:"mode"`
	Video              bool                      `json:"video"`
	Serial             bool                      `json:"serial"`
	SSH                bool                      `json:"ssh"`
	Audio              bool                      `json:"audio"`
	ICEServers         []webrtc.ICEServer        `json:"iceserver"`
	RtcSession         webrtc.SessionDescription `json:"offer" mapstructure:"offer"`
	Describestreamname string                    `json:"streamname"`
	Suuid              string                    `json:"suuid"` //视频流编号，浏览器可以通过预先获取，然后在使用时带过来，主要是提供一个选择分辨率和地址的作用，kvm的话内置4路分辨率，其余的如果是Onvif IPC类则通过Onvif协议在本地获取后通过mqtt传给浏览器，也可以考虑用探测软件实现探测后直接注册到夜莺平台，需要时前端到夜莺平台取
}
type ResponseMsg struct {
	Cmdstr string
	Status int
	Err    string
	Sid    string
}
type PublishMsg struct {
	WEB_SEQID string
	Topic     string
	Msg       interface{}
}
type heartmsg struct {
	Count uint64
}
type handler struct {
	f *os.File
}

func NewHandler() *handler {
	var f *os.File
	if config.Config.Mqtt.WRITETODISK {
		var err error
		f, err = os.Create(config.Config.Mqtt.OUTPUTFILE)
		if err != nil {
			panic(err)
		}
	}
	return &handler{f: f}
}

// Close closes the file
func (o *handler) Close() {
	if o.f != nil {
		if err := o.f.Close(); err != nil {
			fmt.Printf("ERROR closing file: %s", err)
		}
		o.f = nil
	}
}
func SendMsg(msg PublishMsg) {

	msgChans <- msg
	fmt.Print("SendMsg OK")
}

// Add a single video track
func createPeerConnection(msg Message) {
	log.Debug("Incoming CMD message Request")

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers:   msg.ICEServers,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	})
	if err != nil {
		panic(err)
	}
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			// atomic.AddInt64(&peerConnectionCount, -1)
			if err := peerConnection.Close(); err != nil {
				log.Debug("peerConnection.Close error %v", err)
				return
			}
			log.Debug("peerConnection.Closed")
		} else if connectionState == webrtc.ICEConnectionStateConnected {
			log.Debug("peerConnection.Connected ")
			// atomic.AddInt64(&peerConnectionCount, 1)
		}
	})
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		panic(err)
	}
	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		panic(err)
	}

	// audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypePCMA}, "audio", "pion")
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		panic(err)
	}
	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		panic(err)
	}

	offer := msg.RtcSession
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
	<-gatherComplete
	req := &Session{}
	req.Type = CMDMSG_ANSWER
	req.DeviceId = config.Config.Mqtt.CLIENTID

	p := media_interface.Peer{}
	p.InitPeer(msg.Suuid, msg.SeqID, "", "")
	p.AddConnect(msg.Describestreamname, peerConnection)
	p.AddAudioTrack(audioTrack)
	p.AddVideoTrack(videoTrack)
	m := media_interface.GetGlobalStreamM()
	s, err := m.GetStream(msg.Describestreamname)
	if err != nil || s == nil {
		resultstr := fmt.Sprintf("no stream %s", msg.Describestreamname)
		log.Debugf("error %s no stream %s", msg.SeqID, resultstr)
		req.Msg = resultstr
		req.Type = CMDMSG_ERROR
		answermsg := PublishMsg{
			WEB_SEQID: msg.SeqID,
			Topic:     TOPIC_ERROR,
			Msg:       req,
		}
		log.Debugf("error %s", msg.SeqID)
		SendMsg(answermsg)
	} else {
		// s.GetPeer(p.)
		err = s.AddPeer(&p)
		if err != nil {
			resultstr := fmt.Sprintf("no stream %s", msg.Describestreamname)
			log.Debugf("error %s no stream %s", msg.SeqID, resultstr)
			req.Msg = resultstr
			req.Type = CMDMSG_ERROR
			answermsg := PublishMsg{
				WEB_SEQID: msg.SeqID,
				Topic:     TOPIC_ERROR,
				Msg:       req,
			}
			log.Debugf("error %s", msg.SeqID)
			SendMsg(answermsg)
		} else {
			req.Data = enc.Encode(*peerConnection.LocalDescription())
			answermsg := PublishMsg{
				WEB_SEQID: msg.SeqID,
				Topic:     TOPIC_ANSWER,
				Msg:       req,
			}
			log.Debugf("answer %s", msg.SeqID)
			SendMsg(answermsg)
		}
	}
}

func Notice(msg Message) {

	switch msg.Mode {

	case MODE_RTMP:

		go createPeerConnection(msg)

	default:
		answermsg := PublishMsg{
			WEB_SEQID: msg.SeqID,
			Topic:     TOPIC_REFUSE,
			Msg:       "not supported mode" + msg.Mode,
		}
		log.Debugf("answer %s", msg.SeqID)
		SendMsg(answermsg) //response)

	}

}

// handle is called when a message is received
func (o *handler) handle(client mqtt.Client, msg mqtt.Message) {
	// We extract the count and write that out first to simplify checking for missing values
	var m Message
	var resp Session
	if err := json.Unmarshal(msg.Payload(), &resp); err != nil {
		fmt.Printf("Message could not be parsed (%s): %s", msg.Payload(), err)
		return
	}
	log.Debug(resp)
	switch resp.Type {
	case CMDMSG_OFFER:
		enc.Decode(resp.Data, &m)
		Notice(m)
	default:

	}
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Print(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func CmdFeedBack(seqid string, cmdstr string, status int, err string, sid string) {

	resp := ResponseMsg{
		Cmdstr: cmdstr,
		Status: status,
		Err:    err,
		Sid:    sid,
	}
	req := &Session{}
	req.Type = "cmdFeedback"
	req.DeviceId = config.Config.Mqtt.CLIENTID //"kvm1"
	req.Data = enc.Encode(resp)                //enc.Encode(answer)
	answermsg := PublishMsg{
		WEB_SEQID: seqid,
		Topic:     "cmdFeedback",
		Msg:       req,
	}
	log.Debug("cmdFeedback", answermsg)
	SendMsg(answermsg) //response)
}
func GetCurrentPath() string {
	getwd, err := os.Getwd()
	if err != nil {
		fmt.Print(err.Error())
	} else {
		fmt.Print(getwd)
	}
	return getwd
}

func StartMqtt(ctx context.Context) {

	log.Debug("StartMqtt ...")
	// Create a handler that will deal with incoming messages
	h := NewHandler()
	defer h.Close()
	msgChans = make(chan PublishMsg, 10)
	// Now we establish the connection to the mqtt broker
	sn, err := identity.GetSN()
	if err != nil {
		log.Debug("GetSN error", err.Error())
	} else {
		config.Config.Mqtt.CLIENTID = sn
	}

	//只定阅与自身相关的
	config.Config.Mqtt.SUBTOPIC = config.Config.Mqtt.SUBTOPIC + "/" + config.Config.Mqtt.CLIENTID
	config.Config.Mqtt.PUBTOPIC = config.Config.Mqtt.PUBTOPIC + "/" + config.Config.Mqtt.CLIENTID
	log.Debug("subtopic", config.Config.Mqtt.SUBTOPIC, "pubtopic", config.Config.Mqtt.PUBTOPIC)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Config.Mqtt.SERVERADDRESS)
	opts.SetClientID(config.Config.Mqtt.CLIENTID)

	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 30               // Keepalive every 10 seconds so we quickly detect network outages
	opts.PingTimeout = time.Second    // local broker so response should be quick

	// Automate connection management (will keep trying to connect and will reconnect if network drops)
	opts.ConnectRetry = true
	opts.AutoReconnect = true

	// If using QOS2 and CleanSession = FALSE then it is possible that we will receive messages on topics that we
	// have not subscribed to here (if they were previously subscribed to they are part of the session and survive
	// disconnect/reconnect). Adding a DefaultPublishHandler lets us detect this.
	opts.DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("UNEXPECTED MESSAGE: %s\n", msg)
	}

	// Log events
	opts.OnConnectionLost = func(cl mqtt.Client, err error) {
		log.Debug("connection lost")
	}

	opts.OnConnect = func(c mqtt.Client) {
		log.Debug("connection established")

		// Establish the subscription - doing this here means that it willSUB happen every time a connection is established
		// (useful if opts.CleanSession is TRUE or the broker does not reliably store session data)
		t := c.Subscribe(config.Config.Mqtt.SUBTOPIC, config.Config.Mqtt.QOS, h.handle)
		// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
		// in other handlers does cause problems its best to just assume we should not block
		go func() {
			_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
			if t.Error() != nil {
				fmt.Printf("ERROR SUBSCRIBING: %s\n", t.Error())
			} else {
				log.Debug("subscribed to: ", config.Config.Mqtt.SUBTOPIC)
			}
		}()
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Debug("attempting to reconnect")
	}

	//
	// Connect to the broker
	//
	client := mqtt.NewClient(opts)

	// If using QOS2 and CleanSession = FALSE then messages may be transmitted to us before the subscribe completes.
	// Adding routes prior to connecting is a way of ensuring that these messages are processed
	client.AddRoute(config.Config.Mqtt.SUBTOPIC, h.handle)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Debug("Connection is up")
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var count uint64
		for {

			select {
			case data := <-msgChans:
				msg, err := json.Marshal(data.Msg)
				if err != nil {
					panic(err)
				}
				//t := client.Publish(Config.Mqtt.PUBTOPIC+"/"+Config.Report.SN, Config.Mqtt.QOS, false, msg)
				log.Debug("mqtt:", config.Config.Mqtt.PUBTOPIC+"/"+data.WEB_SEQID+"/"+data.Topic)
				t := client.Publish(config.Config.Mqtt.PUBTOPIC+"/"+data.WEB_SEQID+"/"+data.Topic, config.Config.Mqtt.QOS, false, msg)
				go func() {
					_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
					if t.Error() != nil {
						fmt.Printf("msg PUBLISHING: %s\n", t.Error().Error())
					} else {
						//log.Debug("msg PUBLISHING:", msg)
					}
				}()
			case <-time.After(time.Second * time.Duration(config.Config.Mqtt.HEARTTIME)):
				req := &Session{}
				req.Type = "heart"
				req.DeviceId = config.Config.Mqtt.CLIENTID //"kvm1"
				count += 1
				msg, err := json.Marshal(heartmsg{Count: count})
				if err != nil {
					panic(err)
				}
				req.Data = enc.Encode(msg)
				//data := signal.Encode(*peerConnection.LocalDescription())
				answermsg := PublishMsg{
					Topic: "heart",
					Msg:   req,
				}
				msg, err = json.Marshal(answermsg.Msg)
				if err != nil {
					panic(err)
				}
				t := client.Publish(config.Config.Mqtt.PUBTOPIC+"/"+answermsg.Topic, config.Config.Mqtt.QOS, false, msg)
				// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
				go func() {
					_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
					if t.Error() != nil {
						fmt.Printf("ERROR PUBLISHING: %s\n", t.Error().Error())
					} else {
						//log.Debug("HEART PUBLISHING: ", msg)
					}
				}()
			case <-done:
				log.Debug("publisher done")
				wg.Done()
				return
			}
		}
	}()
	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	<-ctx.Done()
	log.Debug("signal caught - exiting")
	client.Disconnect(1000)
	log.Debug("mqtt shutdown complete")
}
