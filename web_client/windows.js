

function FullScreen() {

        var ele = document.getElementById("remote-video");

        if (ele.requestFullscreen) {

            ele.requestFullscreen();
            ele.controls = false;
            ele.controlsList = "nodownload noplaybackrate nofullscreen noremoteplayback";

        } else if (ele.mozRequestFullScreen) {

            ele.mozRequestFullScreen();

        } else if (ele.webkitRequestFullScreen) {

            ele.webkitRequestFullScreen();
            ele.controls = false;
            ele.controlsList = "nodownload noplaybackrate nofullscreen noremoteplayback";

        }

        ele.controls = false;

}

//退出全屏

function exitFullscreen() {

    var de =document.getElementById("remote-video");
    if (de.isFullScreen) {
        return
    }
    if (de.exitFullscreen) {

        de.exitFullscreen();

    } else if (de.mozCancelFullScreen) {

        de.mozCancelFullScreen();

    } else if (de.webkitCancelFullScreen) {

        de.webkitCancelFullScreen();

    }
    return false;
}