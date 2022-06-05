
    // let pc = new RTCPeerConnection()
    // pc.ontrack = function (event) {
    //   var el = document.createElement(event.track.kind)
    //   el.srcObject = event.streams[0]
    //   el.autoplay = true
    //   el.controls = true

    //   document.getElementById('rtmpFeed').appendChild(el)
    // }

    // pc.addTransceiver('video')
    // pc.addTransceiver('audio')
    // pc.createOffer()
    //   .then(offer => {
    //     pc.setLocalDescription(offer)
    //     return fetch(`/createPeerConnection`, {
    //       method: 'post',
    //       headers: {
    //         'Accept': 'application/json, text/plain, */*',
    //         'Content-Type': 'application/json'
    //       },
    //       body: JSON.stringify(offer)
    //     })
    //   })
    //   .then(res => res.json())
    //   .then(res => pc.setRemoteDescription(res))
    //   .catch(alert)

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
function initWebRTC() {
        if (bWebrtc == true) return
        bWebrtc = true
        pc = new RTCPeerConnection({
            iceServers: ICEServer//ICEServer
        });
        //如果是
        if (bVideo) {
            const { receiver } = pc.addTransceiver('video', { direction: 'recvonly' });

            receiver.playoutDelayHint = 0.0;
        }
        if (bAudio) {
            const { receiveraudio } = pc.addTransceiver('audio', { direction: 'recvonly' });
        }
        pc.onsignalingstatechange = ev => {
            switch (pc.signalingState) {
                case "stable":
                    // updateStatus("ICE negotiation complete");
                    break;
            }
        };
        pc.ontrack = function (event) {
            console.log("ontrack", event.track.kind)
            var el = document.createElement(event.track.kind)
            el.srcObject = event.streams[0]
            el.autoplay = true
            el.controls = true
      
            document.getElementById('rtmpFeed').appendChild(el)
        //     if(event.track.kind==="video"){

        //     trackCache = event.track;
        //     var el = document.getElementById('remote-video')
        //     resStream = event.streams[0].clone()
        //     resStream.addTrack(trackCache)
        //     el.srcObject = resStream
           
        //     }else{
        //     var el = document.createElement(event.track.kind);
        //     el.srcObject = event.streams[0];
        //     el.autoplay = true;

        //     document.getElementById("remote-video").appendChild(el);

        //     if (el.nodeName === "AUDIO") {
        //         el.oncanplay = () => {
        //             // el.style = "autoplay"
        //             el.controls = false; // 显示
        //             el.autoplay = true;
        //         };
        //     }
        // }
        }


        pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)


        pc.onicecandidate = event => {
            if (event.candidate === null) {
                //pc.setLocalDescription(offer)
                var msgdata = new Object();
                //var localSessionDescription =btoa(JSON.stringify(pc.localDescription));

                msgdata["seqid"] = WEB_SEQID;
                if (bVideo) {
                    msgdata["video"] = true;
                    msgdata["mode"] = media_mode;
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
                msgdata["iceserver"] = ICEServer;
                msgdata["offer"] = pc.localDescription;//localSessionDescription;
                msgdata["suuid"] = suuid;

                var content = new Object();
                // /content[""]
                content["type"] = CMDMSG_OFFER;
                content["msg"] = "webrtc offer";
                content["device_id"] =document.getElementById("serverId").value //$("#dropdown_menu_link").attr("value");
                content["data"] = btoa(JSON.stringify(msgdata));
                mqttclient.publish(pubtopic, JSON.stringify(content));
                console.log("mqttpublish:",pubtopic, content);
                // console.log("mqttpublish:", btoa(JSON.stringify(content)));
            
                //wsClient.send(JSON.stringify(content));
                // console.log("localDescription:", btoa(JSON.stringify(pc.localDescription)));
            }
        }
        pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
    }
    function endWebrtc() {
        bWebrtc = false;
        var videos = document.getElementsByTagName("remote-video");
        for (var i = 0; i < videos.length; i++) {
            videos[i].pause();
        }
        pc.close();
    }