var pc;
var mqttclient;
var WEB_SEQID;
var suuid;
var MqttServer = "ws://39.98.198.244:8083/mqtt";
var SERVER_NAME="50:9a:4c:3c:2d:b5"//document.getElementById("serverId").value; 与服务器的SN配置结果一致，目前采用的是机器的mac地址
if(SERVER_NAME===""){
    SERVER_NAME="50:9a:4c:3c:2d:b5";
}
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


function log(msg) {
  console.log(msg);
    // $("#loger").html(msg);
}
function startSession() {
  initMqtt();
}
function endSession(){
  endMqtt();
  endWebrtc();
}