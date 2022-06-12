package livekitclient

import (
	"fmt"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/xiangxud/rtmp_webrtc_server/util"
)

type Token struct {
	Create    bool   `yaml:"create" mapstructure:"create"`
	Join      bool   `yaml:"join"  mapstructure:"join"`
	Admin     bool   `yaml:"admin"  mapstructure:"admin"`
	List      bool   `yaml:"list"  mapstructure:"list"`
	Host      string `yaml:"host" mapstructure:"host"`
	ApiKey    string `yaml:"api_key"  mapstructure:"api_key"`
	ApiSecret string `yaml:"api_secret" mapstructure:"api_secret"`
	Identity  string `yaml:"identity"  mapstructure:"identity"`
	RoomName  string `yaml:"room_name"  mapstructure:"room_name"`
	Room      string `yaml:"room"  mapstructure:"room"`
	Metadata  string `yaml:"metadata"  mapstructure:"metadata"`
	ValidFor  string `yaml:"valid_for"  mapstructure:"valid_for"`
}

func (t *Token) GetJoinToken() (string, error) {
	at := auth.NewAccessToken(t.ApiKey, t.ApiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     t.Room,
	}
	at.AddGrant(grant).
		SetIdentity(t.Identity).
		SetValidFor(time.Hour)

	return at.ToJWT()
}

func (t *Token) CreateToken() error {
	if t.ApiKey == "" || t.ApiSecret == "" {
		return fmt.Errorf("api-key and api-secret are required")
	}

	grant := &auth.VideoGrant{}
	if t.Create {
		grant.RoomCreate = true
	}
	if t.Join {
		grant.RoomJoin = true
		grant.Room = t.Room
		if t.Identity == "" {
			return fmt.Errorf("participant identity is required")
		}
	}
	if t.Admin {
		grant.RoomAdmin = true
		grant.Room = t.Room
	}
	if t.List {
		grant.RoomList = true
	}

	if !grant.RoomJoin && !grant.RoomCreate && !grant.RoomAdmin && !grant.RoomList {
		return fmt.Errorf("at least one of --list, --join, --create, or --admin is required")
	}

	at := t.accessToken(grant, t.Identity)

	if t.Metadata != "" {
		at.SetMetadata(t.Metadata)
	}
	if t.RoomName == "" {
		t.RoomName = t.Identity
	}
	at.SetName(t.RoomName)
	if t.ValidFor != "" {
		if dur, err := time.ParseDuration(t.ValidFor); err == nil {
			fmt.Println("valid for (mins): ", int(dur/time.Minute))
			at.SetValidFor(dur)
		} else {
			return err
		}
	}

	token, err := at.ToJWT()
	if err != nil {
		return err
	}

	fmt.Println("token grants")
	util.PrintJSON(grant)
	fmt.Println()
	fmt.Println("access token: ", token)
	return nil
}

func (t *Token) accessToken(grant *auth.VideoGrant, identity string) *auth.AccessToken {

	if t.ApiKey == "" && t.ApiSecret == "" {
		// not provided, don't sign request
		return nil
	}
	at := auth.NewAccessToken(t.ApiKey, t.ApiSecret).
		AddGrant(grant).
		SetIdentity(identity)
	return at
}
