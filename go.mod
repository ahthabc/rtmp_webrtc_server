module github.com/xiangxud/rtmp_webrtc_server

go 1.18

require (
	github.com/Glimesh/go-fdkaac v0.0.0-20220325160929-2f6b0a53a22a
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.4.1
	github.com/livekit/livekit-server v1.1.1
	github.com/livekit/protocol v0.13.5-0.20220706001059-137b876761cb
	github.com/livekit/server-sdk-go v0.10.4-0.20220706190719-2994b147676c
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pion/interceptor v0.1.11
	github.com/pion/ion-sdk-go v0.7.1-0.20220406041327-12e32a5871b9
	github.com/pion/rtcp v1.2.9
	github.com/pion/rtp v1.7.13
	github.com/pion/turn/v2 v2.0.8
	github.com/pion/webrtc/v3 v3.1.42
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.12.0
	github.com/toolkits/pkg v1.3.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/yutopp/go-flv v0.2.0
	github.com/yutopp/go-rtmp v0.0.1
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.21.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/yaml.v3 v3.0.0

)

require (
	github.com/livekit/rtcscore-go v0.0.0-20220524203225-dfd1ba40744a // indirect
	github.com/magefile/mage v1.13.0 // indirect
	github.com/pion/ion-log v1.2.2 // indirect
)

// replace github.com/pion/ion-sdk-go v0.7.0  => github.com/pion/ion-sdk-go.git branch master
require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bep/debounce v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/d5/tengo/v2 v2.10.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/channels v1.1.0 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebml-go/ebml v0.0.0-20160925193348-ca8851a10894 // indirect
	github.com/ebml-go/webm v0.0.0-20160924163542-629e38feef2a // indirect
	github.com/elliotchance/orderedmap v1.4.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/gammazero/deque v0.1.0 // indirect
	github.com/gammazero/workerpool v1.1.2 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-logr/stdr v1.0.0 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-redis/redis/v8 v8.11.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jxskiss/base62 v0.0.0-20191017122030-4f11678b909b // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/mackerelio/go-osstat v0.2.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/petar/GoLLRB v0.0.0-20210522233825-ae3b015fd3e9 // indirect
	github.com/pion/datachannel v1.5.2 // indirect
	github.com/pion/dtls/v2 v2.1.5 // indirect
	github.com/pion/ice/v2 v2.2.6 // indirect
	github.com/pion/ion v1.10.0 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/sctp v1.8.2 // indirect
	github.com/pion/sdp/v3 v3.0.5 // indirect
	github.com/pion/srtp/v2 v2.0.9 // indirect
	github.com/pion/stun v0.3.5 // indirect
	github.com/pion/transport v0.13.1 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sebest/xff v0.0.0-20210106013422-671bd2870b3a // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	github.com/thoas/go-funk v0.9.0 // indirect
	github.com/twitchtv/twirp v8.1.0+incompatible // indirect
	github.com/ua-parser/uap-go v0.0.0-20211112212520-00c877edfe0f // indirect
	github.com/urfave/negroni v1.0.0 // indirect
	github.com/yutopp/go-amf0 v0.0.0-20180803120851-48851794bb1f // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e // indirect
	golang.org/x/sys v0.0.0-20220624220833-87e55d714810 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
