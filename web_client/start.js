
function importScripts(scriptUrl){
  var script= document.createElement("script");
  script.setAttribute("type", "text/javascript");
  script.setAttribute("src", scriptUrl);
  document.body.appendChild(script);
}
importScripts("./parameter.js")
importScripts("./windows.js")
importScripts("./mqtt.js")
importScripts("./webrtc.js")
importScripts("./datachannel.js")
importScripts("./adapter-latest.js")
{/* <script src="https://webrtc.github.io/adapter/adapter-latest.js"></script> */}
function log(msg) {
  console.log(msg);
    // $("#loger").html(msg);
}
function startSession() {
 // endSession()
//  bUseWebrtcP2P= false;
  bUseWebrtcP2P=true;
  bSendCmdMsg=false;
  media_mode = "rtmp";
  subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#";//+"/"+deviceID //Control/00:13:14:01:D9:D5
  pubtopic = "server_control" + "/" + SERVER_NAME;
  initMqtt();
}
function endSession(){
  endMqtt();
  endWebrtc();
}
function stopSession() {
  endMqtt();
  endWebrtc();
}

function startDeviceSession(){
 // endSession()
  bUseWebrtcP2P=true;
  bDevicePull=false;
  WEB_SEQID=uuid();
  media_mode = "offer";
  DEVICE_NAME=document.getElementById("deviceId").value;
  subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#"//+"/"+deviceID //Control/00:13:14:01:D9:D5
  pubtopic = "device_control" + "/" +DEVICE_NAME;
  initMqtt();
}
function startDevicePull(){
  bUseWebrtcP2P= false;
  bDevicePull=true;
  WEB_SEQID=uuid();
  DEVICE_NAME=document.getElementById("deviceId").value;
  subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#"//+"/"+deviceID //Control/00:13:14:01:D9:D5
  pubtopic = "device_control" + "/" + DEVICE_NAME;
  bSendCmdMsg=true;
  initMqtt();
  var msgdata= new Object();
  msgdata["seqid"] = WEB_SEQID;
  if (bVideo) {
      msgdata["video"] = true;
      msgdata["mode"] = "reqpull";//media_mode;
      let name = document.getElementById("streamId");
      let streamname = name.value;
      if (streamname == "") {
          streamname = STREAMNAME;
          name.value = STREAMNAME;
      }
      msgdata["streamname"] = streamname;
      
  }
  if (bAudio) {
      msgdata["audio"] = true;
  }
  var token={
    create: true,
    join: true,
    admin: true,
    list: true,
    host: "ws://192.168.0.18:7880",
    api_key: "APINrg5cyLqPK3p",
    api_secret: "yhmmq0BnW2kTTgGWvwdzwD7MhyEHO5RrDUpprGeBhxe",
    identity: "kvs_device_1",
    room_nane:"kvs_device_1"
  };
  var PullStreamFromDeviceList=[{
    pullfrom: DEVICE_NAME,
    devicename: "kvs"+DEVICE_NAME,
    room_token: [token],
  }];
  msgdata["pull_stream_from_device"] =PullStreamFromDeviceList;
  msgdata["iceserver"] = ICEServer;
  //msgdata["offer"] = pc.localDescription;//localSessionDescription;
  msgdata["suuid"] = suuid;
  msgdata["topicprefix"]=subtopic.substring(0,subtopic.length-2)
  cmd_topic=pubtopic;
  cmd_msgtype="serverpull";
  cmd_deviceid=document.getElementById("serverId").value;
  cmd_msg="req device pull stream";
  cmd_cmdmsg=msgdata;
  // sendCmdMsg(topic,cmdmsgtype,deviceid,msg,cmdmsg);
}
function startDevicetoLivekit(){
  bUseWebrtcP2P= false;
  WEB_SEQID=uuid();
  DEVICE_NAME=document.getElementById("deviceId").value;
  subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#";//+"/"+deviceID //Control/00:13:14:01:D9:D5
  pubtopic = "server_control" + "/" + SERVER_NAME;
  bSendCmdMsg=true;
 
  var msgdata= new Object();
  msgdata["seqid"] = WEB_SEQID;
  if (bVideo) {
      msgdata["video"] = true;
      msgdata["mode"] = "reqpull";//media_mode;
      let name = document.getElementById("streamId");
      let streamname = name.value;
      if (streamname == "") {
          streamname = STREAMNAME;
          name.value = STREAMNAME;
      }
      msgdata["streamname"] = streamname;
      
  }
  if (bAudio) {
      msgdata["audio"] = true;
  }
  var token={
    create: true,
    join: true,
    admin: true,
    list: true,
    host: "ws://192.168.0.18:7880",
    api_key: "APINrg5cyLqPK3p",
    api_secret: "yhmmq0BnW2kTTgGWvwdzwD7MhyEHO5RrDUpprGeBhxe",
    identity: "kvs_device_1",
    room_nane:"kvs_device_1"
  };
  var PullStreamFromDeviceList=[{
    pullfrom: DEVICE_NAME,
    devicename: "kvs"+DEVICE_NAME,
    room_token: [token],
  }];
  msgdata["pull_stream_from_device"] =PullStreamFromDeviceList;
  msgdata["iceserver"] = ICEServer;
  //msgdata["offer"] = pc.localDescription;//localSessionDescription;
  msgdata["suuid"] = suuid;
  cmd_topic=pubtopic;
  cmd_msgtype="serverpull";
  cmd_deviceid=document.getElementById("serverId").value;
  cmd_msg="req device pull stream";
  cmd_cmdmsg=msgdata;
  initMqtt();
  // sendCmdMsg(topic,cmdmsgtype,deviceid,msg,cmdmsg);
}
// type Token struct {
// 	Create    bool   `yaml:"create" mapstructure:"create"`
// 	Join      bool   `yaml:"join"  mapstructure:"join"`
// 	Admin     bool   `yaml:"admin"  mapstructure:"admin"`
// 	List      bool   `yaml:"list"  mapstructure:"list"`
// 	Host      string `yaml:"host" mapstructure:"host"`
// 	ApiKey    string `yaml:"api_key"  mapstructure:"api_key"`
// 	ApiSecret string `yaml:"api_secret" mapstructure:"api_secret"`
// 	Identity  string `yaml:"identity"  mapstructure:"identity"`
// 	RoomName  string `yaml:"room_name"  mapstructure:"room_name"`
// 	Room      string `yaml:"room"  mapstructure:"room"`
// 	Metadata  string `yaml:"metadata"  mapstructure:"metadata"`
// 	ValidFor  string `yaml:"valid_for"  mapstructure:"valid_for"`
// }
// type Pull_Stream_From_Device struct {
// 	Dviceid    string `json:"pullfrom"` //客户端要求服务器拉流的设备标识 也是mqtt topic的唯一值向标识,允许批量
// 	Devicename string `json:"devicename"`
// 	// ICEServers
// 	RoomToken []livekitclient.Token `json:"room_token"` //客户端要求设备加入的房间信息，无此信息则沿用服务器默认的以自身sn创建的房间
// }
// type Message struct {
// 	SeqID                   string                    `json:"seqid"`
// 	Mode                    string                    `json:"mode"`
// 	Pull_Stream_From_Device []Pull_Stream_From_Device `json:"pull_stream_from_device"` //客户端要求的批量拉取设备流的列表
// 	Video                   bool                      `json:"video"`
// 	Serial                  bool                      `json:"serial"`
// 	SSH                     bool                      `json:"ssh"`
// 	Audio                   bool                      `json:"audio"`
// 	ICEServers              []webrtc.ICEServer        `json:"iceserver"`
// 	RtcSession              webrtc.SessionDescription `json:"offer" mapstructure:"offer"`
// 	Describestreamname      string                    `json:"streamname"`
// 	Suuid                   string                    `json:"suuid"` //视频流编号，浏览器可以通过预先获取，然后在使用时带过来，主要是提供一个选择分辨率和地址的作用，kvm的话内置4路分辨率，其余的如果是Onvif IPC类则通过Onvif协议在本地获取后通过mqtt传给浏览器，也可以考虑用探测软件实现探测后直接注册到夜莺平台，需要时前端到夜莺平台取
// }