
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
            var video=document.getElementById('remote-video')
            var el = document.createElement(event.track.kind)
            el.srcObject = event.streams[0]
            el.autoplay = true
            el.controls = true
            document.getElementById('remote-video').appendChild(el)
        }
       initControl();
        pc.oniceconnectionstatechange = e => log(pc.iceConnectionState)


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
        pc.createOffer().then(d => pc.setLocalDescription(d)).catch(log)
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
     
