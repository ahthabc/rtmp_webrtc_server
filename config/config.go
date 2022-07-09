package config

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"github.com/toolkits/pkg/file"
	"github.com/xiangxud/rtmp_webrtc_server/identity"
	"github.com/xiangxud/rtmp_webrtc_server/livekitclient"
	// "github.com/xiangxud/rtmp_webrtc_server/livekitclient"
)

type ConfigT struct {
	HTTP
	Mqtt    mqttSection    `yaml:"mqtt" mapstructure:"mqtt"`
	Stream  streamSection  `yaml:"stream" mapstructure:"stream"`
	Livekit livekitSection `yaml:"livekit" mapstructure:"livekit"`
}
type streamSection struct {
	Debug            bool `yaml:"debug" mapstructure:"debug"`
	IONSfuEnable     bool `yaml:"ionsfu" mapstructure:"ionsfu"`
	LiveKitSfuEnable bool `yaml:"livekitsfu" mapstructure:"livekitsfu"`
}
type livekitSection struct {
	Token livekitclient.Token `yaml:"token"`
	// Config livekitconfig.Config `yaml:"config"`
}
type mqttSection struct {
	SUBTOPIC      string `yaml:"subtopic" mapstructure:"subtopic"` //"topic1"
	PUBTOPIC      string `yaml:"pubtopic" mapstructure:"pubtopic"`
	QOS           byte   `yaml:"qos" mapstructure:"qos"`                     //1
	SERVERADDRESS string `yaml:"serveraddress" mapstructure:"serveraddress"` //= "tcp://mosquitto:1883"
	CLIENTID      string `yaml:"clientid" mapstructure:"clientid"`           //= "mqtt_subscriber"
	WRITETOLOG    bool   `yaml:"writelog" mapstructure:"writelog"`           //= true  // If true then received messages will be written to the console
	WRITETODISK   bool   `yaml:"writetodisk" mapstructure:"writetodisk"`     //= false // If true then received messages will be written to the file below
	OUTPUTFILE    string `yaml:"outputfile" mapstructure:"outputfile"`       //= "/binds/receivedMessages.txt"
	HEARTTIME     int    `yaml:"hearttime" mapstructure:"hearttime"`
	//	CommandLocalPath string `yam:"commanlocalpath"`
}

var (
	Config   ConfigT
	Endpoint string
)

func Parse() error {
	conf := getYmlFile()

	bs, err := file.ReadBytes(conf)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", conf, err)
	}

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(bs))
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", conf, err)
	}

	if err = identity.Parse(); err != nil {
		return err
	}

	var c ConfigT
	err = viper.Unmarshal(&c)
	if err != nil {
		return fmt.Errorf("unmarshal config error:%v", err)
	}
	// fmt.Println("config", c)
	// 启动的时候就获取一下本机的identity，缓存起来以备后用，优点是性能好，缺点是机器唯一标识发生变化需要重启进程
	ident, err := identity.GetIdent()
	if err != nil {
		return err
	}

	fmt.Println("identity:", ident)

	if ident == "" || ident == "127.0.0.1" {
		return fmt.Errorf("identity[%s] invalid", ident)
	}
	c.HTTP = HTTP{ListenAddr: ":8080", CORS: true, mux: http.DefaultServeMux}
	Endpoint = ident
	Config = c
	sn, err := identity.GetSN()
	if err == nil {
		Config.Livekit.Token.Identity = sn
	}
	return nil
}

func getYmlFile() string {
	yml := "etc/serverconf.local.yml"
	if file.IsExist(yml) {
		return yml
	}

	yml = "./etc/serverconf.yml"
	if file.IsExist(yml) {
		return yml
	}

	return ""
}
