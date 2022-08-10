

// function importScript(scriptUrl){
//     var script= document.createElement("script");
//     script.setAttribute("type", "text/javascript");
//     script.setAttribute("src", scriptUrl);
//     document.body.appendChild(script);
// }


var t = Date.now();
 
function sleep(d){
	while(Date.now - t <= d);
}
function initMqtt() {
    if(bmqttStarted){
        console.log("mqtt is connect");
        return;
    }
    var ClientId = 'mqttjs_' + Math.random().toString(16).substr(2, 8)
    // var options = {
    //     invocationContext: {
    //         host: hostname,
    //         port: port,
    //         path: client.path,
    //         clientId: clientId
    //     },
    //     timeout: timeout,
    //     keepAliveInterval: keepAlive,
    //     cleanSession: cleanSession,
    //     useSSL: ssl,//wss传输
    //     userName: userName,  
    //     password: password,  
    //     onSuccess: onConnect,
    //     mqttVersion: 4,
    //     onFailure: function (e) {
    //         console.log(e);
    //         s = "{time:" + new Date().Format("yyyy-MM-dd hh:mm:ss") + ", onFailure()}";
    //         console.log(s);
    //     }
    // };
    mqttclient = mqtt.connect(MqttServer,
        {
            clientId: ClientId,
            // useSSL: true,
            // protocol: "wss",
            // rejectUnauthorized: false,
            // ca: 'CA signed server certificate',
            username: 'admin',
            password: 'password',
            // port: 8084
        });
    mqttclient.on('connect', function () {
        mqttclient.subscribe(subtopic, function (err) {
            if (!err) {
                //mqttclient.publish('Control', 'Hello mqtt')
                //成功连接到服务器
                console.log("connected to server");
                bmqttStarted=true;
                if(bUseWebrtcP2P)
                   initWebRTC();
                if(bSendCmdMsg){
                    sendCmdMsg(cmd_topic,cmd_msgtype,cmd_deviceid,cmd_msg,cmd_cmdmsg);
                }   
            }
        })
    })
    mqttclient.on('message', function (topic, message) {
        // message is Buffer
        console.log("topic:",topic)
        console.log("message:",message)

        let input = JSON.parse(message)
        console.log("input:",input)
        switch (input.type) {
            case 'offer': 
              getRemoteOffer(input);
              break;
            case "error":
                console.log("msg:",input.msg);
                stopSession();
                break;
            case "answer":
                var remoteSessionDescription = input.data;
                if (remoteSessionDescription === '') {
                    alert('Session Description must not be empty');
                }
                try {
                    let answerjsonstr=atob(remoteSessionDescription);
                    console.log("atob1:",answerjsonstr);
                    // answerjsonstr=answerjsonstr.substring(0,answerjsonstr.length-1);
                    // console.log("atob2:",answerjsonstr);

                    // let answer1 = answerjsonstr
                    let answer = JSON.parse(answerjsonstr);
                    console.log("answer:",answer);
                    // let answerjson=new RTCSessionDescription(answer);
                    // pc.setRemoteDescription(printAndReturnRemoteDescription(answer));
                    // pc.setRemoteDescription(printAndReturnRemoteDescription(answer));
                    pc.setRemoteDescription(new RTCSessionDescription(answer));
                    // btnOpen();
                } catch (e) {
                    alert(e);
                }
                break;
            // case CMDMSG_DISCRSP:
            //     console.log(JSON.parse(atob(input.data)));
            //     getDevices(JSON.parse(atob(input.data)))
            //     break;
            case "heart":
                console.log(JSON.parse(atob(input.data)));
                break;
            case "cmdFeedback":
                console.log(JSON.parse(atob(input.data)));
                break;
            // case CMDMSG_PROCLIST:
            //     console.log(JSON.parse(atob(input.data)));
            //     break;
            // case CMDMSG_RESPKVMRTSPINFOLIST:
            //     console.log(JSON.parse(atob(input.data)));
            //     break;
        }
    })
}
function endMqtt() {
    if(!bmqttStarted) return;
    mqttclient.end()
    bmqttStarted=false;
}

function sendCmdMsg(topic,cmdmsgtype,deviceid,msg,cmdmsg){
    // var msgdata = new Object();
    //var localSessionDescription =btoa(JSON.stringify(pc.localDescription));
    // while(!bmqttStarted){
    //     console.log("mqtt 没链接");
    //     sleep(1);
    // }
    // var timer = setInterval(function () {
    // if(!bmqttStarted){
    //     console.log("mqtt 没链接");
    //     return;
    // }
    // clearTimeout(timer); //关闭定时器。
    var content = new Object();
    // /content[""]
    content["type"] = cmdmsgtype;//CMDMSG_OFFER;
    content["msg"] = msg;//"webrtc offer";
    content["device_id"] =deviceid;//document.getElementById("serverId").value //$("#dropdown_menu_link").attr("value");
    content["data"] = btoa(JSON.stringify(cmdmsg));
    mqttclient.publish(topic, JSON.stringify(content));
    console.log("mqttpublish:",topic, cmdmsg);
    // },1000);

    // console.log("mqttpublish:", btoa(JSON.stringify(content)));

    //wsClient.send(JSON.stringify(content));
    // console.log("localDescription:", btoa(JSON.stringify(pc.localDescription)));
}
