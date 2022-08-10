var pc;
var mqttclient;
var WEB_SEQID;
var suuid;
var local;
var localStream;
// var MqttServerWSS = "wss://192.168.0.18:8084/mqtt"//"wss://192.168.0.18:8084/mqtt";
// var MqttServer = "wss://39.98.198.244:8084/mqtt"
var MqttServer="wss://192.168.0.18:8084/mqtt";
var SERVER_NAME="50:9a:4c:3c:2d:b5"//document.getElementById("serverId").value; 与服务器的SN配置结果一致，目前采用的是机器的mac地址
var DEVICE_NAME="50:9A:4C:3C:2D:B5"//"4E:7B:BF:71:D6:B4"
var kvs=true;
if(SERVER_NAME===""){
    SERVER_NAME="50:9a:4c:3c:2d:b5";
}
let startTime;
var remoteVideo=document.getElementById('remote-video');

// -------- codec 的配置 --------
const codecPreferences = document.getElementById('codecPreferences');
const supportsSetCodecPreferences = window.RTCRtpTransceiver &&
  'setCodecPreferences' in window.RTCRtpTransceiver.prototype;
// -----------------------------
  if (supportsSetCodecPreferences) {
    const { codecs } = RTCRtpSender.getCapabilities('video');
    console.log('RTCRtpSender.getCapabilities(video):\n', codecs);
    codecs.forEach(codec => {
      if (['video/red', 'video/ulpfec', 'video/rtx'].includes(codec.mimeType)) {
        return;
      }
      const option = document.createElement('option');
      option.value = (codec.mimeType + ' ' + (codec.sdpFmtpLine || '')).trim();
      option.innerText = option.value;
      codecPreferences.appendChild(option);
    });
    codecPreferences.disabled = false;
  } else {
    console.warn('当前不支持更换codec');
  }
function suuid() {
     
	var s = [];
	var hexDigits = "0123456789abcdef";
	for (var i = 0; i < 36; i++) {
     
		s[i] = hexDigits.substr(Math.floor(Math.random() * 0x10), 1);
	}
	s[14] = "4"; // bits 12-15 of the time_hi_and_version field to 0010
	s[19] = hexDigits.substr((s[19] & 0x3) | 0x8, 1); // bits 6-7 of the clock_seq_hi_and_reserved to 01
	s[8] = s[13] = s[18] = s[23] = "-";
	
	var uuid1 = s.join("");
	return uuid1;
}
function uuid() {
    var temp_url = URL.createObjectURL(new Blob());
    var uuid = temp_url.toString(); // blob:https://xxx.com/b250d159-e1b6-4a87-9002-885d90033be3
    URL.revokeObjectURL(temp_url);
    return uuid.substr(uuid.lastIndexOf("/") + 1);
}
 WEB_SEQID=uuid();
 suuid=suuid();
var subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#";//+"/"+deviceID //Control/00:13:14:01:D9:D5
var pubtopic = "server_control" + "/" + SERVER_NAME;
let bVideo=true;
let bAudio=true;
var bmqttStarted=false; 
var bWebrtc = false;
var bUseWebrtcP2P =true;//启动webrtc p2p 模式
var bSendCmdMsg = false;
var bUseMesg=false; //发送cmd msg 
var bDevicePull=false; //设备推流 true 客户端拉流false
var cmd_topic;
var cmd_msgtype;
var cmd_deviceid;
var cmd_msg;
var cmd_cmdmsg;
var controlDC;
var bcontrolopen = false;
const CMDMSG_OFFER = "offer"
const CMDMSG_ANSWER = "answer"
var STREAMNAME=document.getElementById("streamId").value;
if(STREAMNAME===""){
    STREAMNAME="kvs";
}
let media_mode = "rtmp";
// var ICEServer ={iceServers: [
//     {
//         urls: ["stun:192.168.0.18:3478"]
//         // url: "stun:39.98.198.244:3478"
//         //url:"stun:stun.l.google.com:19302"

//     }, {
//         urls: ["turn:192.168.0.18:3478"],
//         // url: "turn:39.98.198.244:3478",
//         username: "media",
//         credential: "123456"
//     }
// ], sdpSemantics:'plan-b'};
var ICEServer =[
    {
        urls: ["stun:192.168.0.18:3478"]
        // url: "stun:39.98.198.244:3478"
        //url:"stun:stun.l.google.com:19302"

    }, {
        urls: ["turn:192.168.0.18:3478"],
        // url: "turn:39.98.198.244:3478",
        username: "media",
        credential: "123456"
    }
];
// var ICEServer = [
//         {
//             urls: "stun:stun.l.google.com:19302",
//             // url: "stun:39.98.198.244:3478"
//             //url:"stun:stun.l.google.com:19302"
    
//         }
//     ];