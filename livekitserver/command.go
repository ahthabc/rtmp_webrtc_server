package livekitserver

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/utils"

	"github.com/livekit/livekit-server/pkg/routing"
	"github.com/livekit/livekit-server/pkg/service"
)

func generateKeys(_ *cli.Context) error {
	apiKey := utils.NewGuid(utils.APIKeyPrefix)
	secret := utils.RandomSecret()
	fmt.Println("API Key: ", apiKey)
	fmt.Println("API Secret: ", secret)
	return nil
}

func printPorts(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	udpPorts := make([]string, 0)
	tcpPorts := make([]string, 0)

	tcpPorts = append(tcpPorts, fmt.Sprintf("%d - HTTP service", conf.Port))
	if conf.RTC.TCPPort != 0 {
		tcpPorts = append(tcpPorts, fmt.Sprintf("%d - ICE/TCP", conf.RTC.TCPPort))
	}
	if conf.RTC.UDPPort != 0 {
		udpPorts = append(udpPorts, fmt.Sprintf("%d - ICE/UDP", conf.RTC.UDPPort))
	} else {
		udpPorts = append(udpPorts, fmt.Sprintf("%d-%d - ICE/UDP range", conf.RTC.ICEPortRangeStart, conf.RTC.ICEPortRangeEnd))
	}

	if conf.TURN.Enabled {
		if conf.TURN.TLSPort > 0 {
			tcpPorts = append(tcpPorts, fmt.Sprintf("%d - TURN/TLS", conf.TURN.TLSPort))
		}
		if conf.TURN.UDPPort > 0 {
			udpPorts = append(udpPorts, fmt.Sprintf("%d - TURN/UDP", conf.TURN.UDPPort))
		}
	}

	fmt.Println("TCP Ports")
	for _, p := range tcpPorts {
		fmt.Println(p)
	}

	fmt.Println("UDP Ports")
	for _, p := range udpPorts {
		fmt.Println(p)
	}
	return nil
}

func createToken(c *cli.Context) error {
	room := c.String("room")
	identity := c.String("identity")

	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	// use the first API key from config
	if len(conf.Keys) == 0 {
		// try to load from file
		if _, err := os.Stat(conf.KeyFile); err != nil {
			return err
		}
		f, err := os.Open(conf.KeyFile)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		decoder := yaml.NewDecoder(f)
		if err = decoder.Decode(conf.Keys); err != nil {
			return err
		}

		if len(conf.Keys) == 0 {
			return fmt.Errorf("keys are not configured")
		}
	}

	var apiKey string
	var apiSecret string
	for k, v := range conf.Keys {
		apiKey = k
		apiSecret = v
		break
	}

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	if c.Bool("recorder") {
		grant.Hidden = true
		grant.Recorder = true
		grant.SetCanPublish(false)
		grant.SetCanPublishData(false)
	}

	at := auth.NewAccessToken(apiKey, apiSecret).
		AddGrant(grant).
		SetIdentity(identity).
		SetValidFor(30 * 24 * time.Hour)

	token, err := at.ToJWT()
	if err != nil {
		return err
	}

	fmt.Println("Token:", token)

	return nil
}

func listNodes(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	currentNode, err := routing.NewLocalNode(conf)
	if err != nil {
		return err
	}

	router, err := service.InitializeRouter(conf, currentNode)
	if err != nil {
		return err
	}

	nodes, err := router.ListNodes()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{
		"ID", "IP Address", "Region",
		"CPUs", "CPU Usage /\nLoad Avg",
		"Rooms", "Clients /\nTracks In/Out",
		"Bytes/s In/Out /\nBytes Total", "Packets/s In/Out /\nPackets Total",
		"Nack/s /\nNack Total", "Retrans/s /\nRetrans Total",
		"Started At /\nUpdated At",
	})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER,
		tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_CENTER,
	})

	for _, node := range nodes {
		stats := node.Stats

		// Id and state
		idAndState := fmt.Sprintf("%s\n(%s)", node.Id, node.State.Enum().String())

		// System stats
		cpus := strconv.Itoa(int(stats.NumCpus))
		cpuUsageAndLoadAvg := fmt.Sprintf("%.2f %%\n%.2f %.2f %.2f", stats.CpuLoad*100,
			stats.LoadAvgLast1Min, stats.LoadAvgLast5Min, stats.LoadAvgLast15Min)

		// Room stats
		rooms := strconv.Itoa(int(stats.NumRooms))
		clientsAndTracks := fmt.Sprintf("%d\n%d / %d", stats.NumClients, stats.NumTracksIn, stats.NumTracksOut)

		// Packet stats
		bytes := fmt.Sprintf("%sps / %sps\n%s / %s", humanize.Bytes(uint64(stats.BytesInPerSec)), humanize.Bytes(uint64(stats.BytesOutPerSec)),
			humanize.Bytes(stats.BytesIn), humanize.Bytes(stats.BytesOut))
		packets := fmt.Sprintf("%s / %s\n%s / %s", humanize.Comma(int64(stats.PacketsInPerSec)), humanize.Comma(int64(stats.PacketsOutPerSec)),
			strings.TrimSpace(humanize.SIWithDigits(float64(stats.PacketsIn), 2, "")), strings.TrimSpace(humanize.SIWithDigits(float64(stats.PacketsOut), 2, "")))
		nacks := fmt.Sprintf("%.2f\n%s", stats.NackPerSec, strings.TrimSpace(humanize.SIWithDigits(float64(stats.NackTotal), 2, "")))
		retransmit := fmt.Sprintf("%.2f\n%s", stats.RetransmitPacketsOutPerSec, strings.TrimSpace(humanize.SIWithDigits(float64(stats.RetransmitPacketsOut), 2, "")))

		// Date
		startedAndUpdated := fmt.Sprintf("%s\n%s", time.Unix(stats.StartedAt, 0).UTC().UTC().Format("2006-01-02 15:04:05"),
			time.Unix(stats.UpdatedAt, 0).UTC().Format("2006-01-02 15:04:05"))

		table.Append([]string{
			idAndState, node.Ip, node.Region,
			cpus, cpuUsageAndLoadAvg,
			rooms, clientsAndTracks,
			bytes, packets,
			nacks, retransmit,
			startedAndUpdated,
		})
	}
	table.Render()

	return nil
}
