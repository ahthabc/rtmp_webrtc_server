var pc;
var mqttclient;
var WEB_SEQID;
var suuid;
var MqttServer = "ws://192.168.0.18:8083/mqtt";
var SERVER_NAME="50:9a:4c:3c:2d:b5"//document.getElementById("serverId").value; 与服务器的SN配置结果一致，目前采用的是机器的mac地址
if(SERVER_NAME===""){
    SERVER_NAME="50:9a:4c:3c:2d:b5";
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
var bWebrtc = false;
const CMDMSG_OFFER = "offer"
var STREAMNAME=document.getElementById("streamId").value;
if(STREAMNAME===""){
    STREAMNAME="test";
}
let media_mode = "rtmp";
var ICEServer = [
    {
        url: "stun:192.168.0.18:3478"
        // url: "stun:39.98.198.244:3478"
        //url:"stun:stun.l.google.com:19302"

    }, {
        url: "turn:192.168.0.18:3478",
        // url: "turn:39.98.198.244:3478",
        username: "media",
        credential: "123456"
    }
];
