
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

function setkvstype(){
    kvs=true;
    console.log("Setting kvsRTC");
}
function setmetartctype(){
    kvs=false;
    console.log("Setting metaRTC");
}

function gotStream(stream){
  localvideo.src = webkitURL.createObjectURL(stream);
}
function getUserMedia(obj,success,error){
    if(navigator.getUserMedia){
    getUserMedia=function(obj,success,error){
    navigator.getUserMedia(obj,function(stream){
    success(stream);
    },error);
    }
    }else if(navigator.webkitGetUserMedia){
    getUserMedia=function(obj,success,error){
    navigator.webkitGetUserMedia(obj,function(stream){
    var _URL=window.URL || window.webkitURL;
    success(_URL.createObjectURL(stream));
    },error);
    }
    }else if(navigator.mozGetUserMedia){
    getUserMedia=function(obj,success,error){
    navigator.mozGetUserMedia(obj,function(stream){
    success(window.URL.createObjectURL(stream));
    },error);
    }
    }else{
    return false;
    }
    return getUserMedia(obj,success,error);
}
async function onIceCandidate(pc, event) {
    try {
    //   await (getOtherPc(pc).addIceCandidate(event.candidate));
      onAddIceCandidateSuccess(pc);
    } catch (e) {
      onAddIceCandidateError(pc, e);
    }
    console.log(`${getName(pc)} ICE candidate:\n${event.candidate ? event.candidate.candidate : '(null)'}`);
  }

  function onAddIceCandidateSuccess(pc) {
    console.log(`${getName(pc)} addIceCandidate success`);
  }
  function onAddIceCandidateError(pc, error) {
    console.log(`${getName(pc)} failed to add ICE Candidate: ${error.toString()}`);
  }  

  function getName(pc) {
    return (pc === pc) ?'pc':'pc';//? 'pc1' : 'pc2';
  }
//   https://www.bbsmax.com/A/lk5a8R7PJ1/
// https://blog.csdn.net/lym594887256/article/details/124472804
// https://blog.csdn.net/ice_ly000/article/details/105763753
function  initWebRTC()  {
        startTime = window.performance.now();
        if (bWebrtc == true) return
        bWebrtc = true
        pc = new RTCPeerConnection({
            iceServers: ICEServer//ICEServer
        });
        const videoTracks = localStream.getVideoTracks();
        const audioTracks = localStream.getAudioTracks();
        if (videoTracks.length > 0) {
          console.log(`使用的摄像头: ${videoTracks[0].label}`);
        }
        if (audioTracks.length > 0) {
          console.log(`使用的麦克风: ${audioTracks[0].label}`);
        }
        // const configuration = {};
        // pc1 = new RTCPeerConnection(configuration);
        // pc1.addEventListener('icecandidate', e => onIceCandidate(pc1, e));
        // pc2 = new RTCPeerConnection(configuration);
        // pc2.addEventListener('icecandidate', e => onIceCandidate(pc2, e));
        // pc2.addEventListener('track', gotRemoteStream);    

        pc.addEventListener('icecandidate', e => onIceCandidate(pc, e));
        var radio=document.getElementsByName("apptype");

        for(var i=0;i<radio.length;i++){

           if(radio[i].checked){
             if(i==0) setkvstype();
             else setmetartctype();
        //   radio[i].addEventListener("click",clickFunction);

          }

        }
        
        // const videotrack,audiotrack;
        // codec_parameters = OrderedDict(
        //     [
        //         ("packetization-mode", "1"),
        //         ("level-asymmetry-allowed", "1"),
        //         ("profile-level-id", "42001f"),
        //     ]
        // )
        
        // h264_capability = RTCRtpCodecCapability(
        //     mimeType="video/H264", clockRate=90000, channels=None, parameters=codec_parameters
        // )
        
        // preferences = [h264_capability]
        // later to be applied to the video transceiver
        
        // for t in pc.getTransceivers():
        //         if t.kind == "audio" and player.audio:
        //             pc.addTrack(player.audio)
        //         elif t.kind == "video" and player.video:
        //             pc.addTrack(player.video)
        //             t.setCodecPreferences(preferences)

        //如果是
        //     if(kvs){
        //         console.log(" kvsRTC mode");
        //         if (bVideo) {

        //             for (const track of localStream.getTracks()) {
        //                 if(track.kind=="vedio")  {pc.addTrack(track); break};
        //                 // else if(track.kind=="audioinput") audiotrack=track;
        //                 // peerconnetion.addTrack(track);
        //             }
                   
        //             // const { receiver } =  pc.addTransceiver('video', { direction: 'sendrecv' });
        
        //             //  receiver.playoutDelayHint = 0.0;
        //         }
        //         if (bAudio) {
        //             for (const track of localStream.getTracks()) {
        //                 if(track.kind=="audio")  {pc.addTrack(track); break};
        //                 // else if(track.kind=="audioinput") audiotrack=track;
        //                 // peerconnetion.addTrack(track);
        //             }
        //             // pc.addTrack(audiotrack);
        //             // const { receiveraudio } = pc.addTransceiver('audio', { direction: 'sendrecv' });
        //         }
        
        //     }else{
        //         console.log(" metaRTC mode");
        //         if (bAudio) {
        //             for (const track of localStream.getTracks()) {
        //                 if(track.kind=="vedio")  {pc.addTrack(track); break};
        //                 // else if(track.kind=="audioinput") audiotrack=track;
        //                 // peerconnetion.addTrack(track);
        //             }
        //             // pc.addTrack(audiotrack);
        //             // const { receiveraudio } = pc.addTransceiver('audio', { direction: 'sendrecv' });
        //         }
        //         if (bVideo) {
        //             for (const track of localStream.getTracks()) {
        //                 if(track.kind=="audio")  {pc.addTrack(track); break};
        //                 // else if(track.kind=="audioinput") audiotrack=track;
        //                 // peerconnetion.addTrack(track);
        //             }
        //             // pc.addTrack(videotrack);
        //             // const { receiver } = pc.addTransceiver('video', { direction: 'sendrecv' });

        //             // receiver.playoutDelayHint = 0.0;
        //         }
        //   }
        //   for (const t of pc.getTransceivers()){
        //         if(t.kind == "video"){
        //             t.setCodecPreferences(preferences)
        //         }
        // }
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
            document.getElementById('remote-video').appendChild(el)
        }


       initControl();
       localStream.getTracks().forEach(track => pc.addTrack(track, localStream));
        pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)
        if (supportsSetCodecPreferences) {
            // 获取选择的codec
            const preferredCodec = codecPreferences.options[codecPreferences.selectedIndex];
            if (preferredCodec.value !== '') {
              const [mimeType, sdpFmtpLine] = preferredCodec.value.split(' ');
              const { codecs } = RTCRtpSender.getCapabilities('video');
              const selectedCodecIndex = codecs.findIndex(c => c.mimeType === mimeType && c.sdpFmtpLine === sdpFmtpLine);
              const selectedCodec = codecs[selectedCodecIndex];
              codecs.splice(selectedCodecIndex, 1);
              codecs.unshift(selectedCodec);
              console.log(codecs);
              const transceiver = pc.getTransceivers().find(t => t.sender && t.sender.track === localStream.getVideoTracks()[0]);
              transceiver.setCodecPreferences(codecs);
              console.log('选择的codec', selectedCodec);
            }
          }
        codecPreferences.disabled = true;
        remoteVideo.addEventListener('resize', () => {
        console.log(`Remote video size changed to ${remoteVideo.videoWidth}x${remoteVideo.videoHeight}`);
        if (startTime) {
            const elapsedTime = window.performance.now() - startTime;
            console.log('视频流连接耗时: ' + elapsedTime.toFixed(3) + 'ms');
            startTime = null;
        }
        });
        pc.onicecandidate = event => {
            if(!bDevicePull){
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
                msgdata["topicprefix"]=subtopic.substring(0,subtopic.length-2)
                var content = new Object();
                // /content[""]
                content["type"] = CMDMSG_OFFER;
                content["msg"] = "webrtc offer";
                content["device_id"] =document.getElementById("serverId").value //$("#dropdown_menu_link").attr("value");
                content["data"] = btoa(JSON.stringify(msgdata));
                mqttclient.publish(pubtopic, JSON.stringify(content));
                console.log("mqttpublish:",pubtopic, msgdata);
                // console.log("mqttpublish:", btoa(JSON.stringify(content)));
            
                //wsClient.send(JSON.stringify(content));
                // console.log("localDescription:", btoa(JSON.stringify(pc.localDescription)));
            }
        }
        }
        if(!bDevicePull){
        const offerOption = {
            offerToReceiveAudio: true,
            offerToReceiveVideo: true,
        };
        pc.createOffer(offerOption).then(d => pc.setLocalDescription(d)).catch(log)
        }
    }
    function endWebrtc() {
        bWebrtc = false;
        pc.close();
        var videos = document.getElementById("remote-video");
        var len=videos.childNodes.length
        for (var i = 0; i < len; i++) {
            videos.removeChild(videos.childNodes[0])
        }
    }

function getRemoteOffer(input){
        var remoteSessionDescription = input.data;
        if (remoteSessionDescription === '') {
            alert('Session Description must not be empty');
        }
        try {
            let offermsgjsonstr=atob(remoteSessionDescription);
            console.log("atob1:",offermsgjsonstr);
            // answerjsonstr=answerjsonstr.substring(0,answerjsonstr.length-1);
            // console.log("atob2:",answerjsonstr);

            // let answer1 = answerjsonstr
            let msg = JSON.parse(offermsgjsonstr);
            console.log("msg:",msg);
            // let offerjsonstr=atob(msg["offer"])
            let offer=msg["offer"];//JSON.parse(offerjsonstr)
            let iceserver=msg["iceserver"];//JSON.parse(iceserver)
            let offersdp=new RTCSessionDescription(offer);
            console.log("offer",offer,offersdp);
            if (bWebrtc == true) return
            bWebrtc = true
            pc = new RTCPeerConnection({
                iceServers: iceserver//ICEServer
            });
            
            // pc.setRemoteDescription(offersdp);
            pc.setRemoteDescription(offersdp);
            pc.createAnswer()
            .then(sdp => pc.setLocalDescription(sdp)).then(() => {
                 gotDescription(pc.localDescription);    //   socket.emit("answer", id, peerConnection.localDescription);
                });
            // pc.setLocalDescription();
            // .then(() => pc.createAnswer())
            // .then(sdp => pc.setLocalDescription(sdp))
            // .then(() => {
            
            //     gotDescription(pc.localDescription);    //   socket.emit("answer", id, peerConnection.localDescription);
            // });



            pc.onsignalingstatechange = ev => {
                switch (pc.signalingState) {
                    case "stable":
                        console.log("pc.signalingState is stable")
                        // updateStatus("ICE negotiation complete");
                        break;
                }
            };
            pc.ontrack = function (event) {
                console.log("ontrack", event.track.kind)
                var video=document.getElementById('remote-video')
                var el = document.createElement(event.track.kind)
                el.srcObject = event.streams[0]
                el.autoplay = true
                el.controls = true
                document.getElementById('remote-video').appendChild(el)
            }
            pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)


            pc.onicecandidate = event => {
                if (event.candidate != null){
                 console.log("onicecandidate ",event.candidate);
                }
            }
            //  pc.setLocalDescription().catch(log);

        } catch (e) {
            alert(e);
        }



        //     console.log("answer",answer,"remote Description",pc.remoteDescription,"localDescription",pc.localDescription)
           
        //     gotDescription(answer);
        // }).catch(log);
        // pc.createAnswer(pc.remoteDescription, gotDescription);
        function gotDescription(desc,candidate,mode_type) {
            pc.setLocalDescription(desc).catch(log);
            var msgdata = new Object();
            msgdata["seqid"] = WEB_SEQID;
            msgdata["mode"] = CMDMSG_ANSWER;
            msgdata["answer"] = pc.localDescription;//localSessionDescription;
            msgdata["suuid"] = suuid;
            msgdata["candidate"]=candidate;
            msgdata["topicprefix"]=subtopic.substring(0,subtopic.length-2)
            var content = new Object();
            // /content[""]
            content["type"] = CMDMSG_ANSWER;
            content["msg"] = "webrtc answer";
            content["device_id"] =document.getElementById("serverId").value //$("#dropdown_menu_link").attr("value");
            content["data"] = btoa(JSON.stringify(msgdata));
            mqttclient.publish(pubtopic, JSON.stringify(content));
            console.log("mqttpublish:",pubtopic, msgdata);
    
            // signalingChannel.send(JSON.stringify({ "sdp": desc }));
        }
        // pc.createAnswer().then(d => pc.setLocalDescription(d)).catch(log)

}
     
