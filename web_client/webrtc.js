
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
            var el = document.createElement(event.track.kind)
            el.srcObject = event.streams[0]
            el.autoplay = true
            el.controls = true
      
            document.getElementById('remote-video').appendChild(el)
        //     if(event.track.kind==="video"){

        //         trackCache = event.track;
        //         var el = document.getElementById('remote-video')
        //         resStream = event.streams[0].clone()
        //         resStream.addTrack(trackCache)
        //         el.srcObject = resStream
           
        //     }else{
        //         var el = document.createElement(event.track.kind);
        //         el.srcObject = event.streams[0];
        //         el.autoplay = true;

        //         document.getElementById("remote-video").appendChild(el);

        //         if (el.nodeName === "AUDIO") {
        //             el.oncanplay = () => {
        //                 // el.style = "autoplay"
        //                 el.controls = false; // 显示
        //                 el.autoplay = true;
        //             };
        //         }
        
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