
var SERVER_NAME="50:9a:4c:3c:2d:b5"//document.getElementById("serverId").value;
if(SERVER_NAME===""){
    SERVER_NAME="50:9a:4c:3c:2d:b5";
}
var subtopic = "server_cmd/" +SERVER_NAME+ "/"+ WEB_SEQID + "/#";//+"/"+deviceID //Control/00:13:14:01:D9:D5
var pubtopic = "server_control" + "/" + SERVER_NAME;
function initMqtt() {

    var ClientId = 'mqttjs_' + Math.random().toString(16).substr(2, 8)
    mqttclient = mqtt.connect(MqttServer,
        {
            clientId: ClientId,
            username: 'admin',
            password: 'password'
        });
    mqttclient.on('connect', function () {
        mqttclient.subscribe(subtopic, function (err) {
            if (!err) {
                //mqttclient.publish('Control', 'Hello mqtt')
                //成功连接到服务器
                console.log("connected to server");
                initWebRTC();
            }
        })
    })
    mqttclient.on('message', function (topic, message) {
        // message is Buffer
        console.log(topic)
        console.log(message)

        let input = JSON.parse(message)
        console.log(input)
        switch (input.type) {
            case "error":
                console.log(input.msg);
                stopSession();
                break;
            case "answer":
                var remoteSessionDescription = input.data;
                if (remoteSessionDescription === '') {
                    alert('Session Description must not be empty');
                }
                try {
                    let answer = JSON.parse(atob(remoteSessionDescription));
                    console.log(answer);
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
    mqttclient.end()
}