
function importScripts(scriptUrl){
  var script= document.createElement("script");
  script.setAttribute("type", "text/javascript");
  script.setAttribute("src", scriptUrl);
  document.body.appendChild(script);
}
importScripts("./parameter.js")
importScripts("./mqtt.js")
importScripts("./webrtc.js")

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
function stopSession() {
  endMqtt();
  endWebrtc();
}