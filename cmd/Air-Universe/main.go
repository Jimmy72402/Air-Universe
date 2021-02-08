package main

import (
	"flag"
	"fmt"
	v2rayApi "github.com/crossfw/Air-Universe/pkg/V2rayApi"
	"github.com/crossfw/Air-Universe/pkg/structures"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
	_ "time"
)

const (
	VERSION = "0.0.5"
)

func init() {
	log.SetLevel(log.DebugLevel)

	var (
		printVersion bool
		configPath   string
	)

	flag.BoolVar(&printVersion, "V", false, "print version")
	flag.StringVar(&configPath, "C", "config/Air-Universe_json/test.json", "configure file")
	flag.Parse()

	if printVersion {
		_, _ = fmt.Fprintf(os.Stdout, "V2ray-ssp %s\n", VERSION)
		os.Exit(0)
	}

	if configPath != "" {
		_, err := parseBaseConfig(&configPath)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		err = checkCfg()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	//if flag.NFlag() == 0 {
	//	flag.PrintDefaults()
	//	os.Exit(0)
	//}

}

func sync(idIndex uint32) error {
	var (
		err                   error
		v2Client              v2rayController
		usersBefore, usersNow *[]structures.UserInfo
		usersTraffic          *[]structures.UserTraffic
	)
	usersBefore = new([]structures.UserInfo)
	usersNow = new([]structures.UserInfo)
	usersTraffic = new([]structures.UserTraffic)

	// Get gRpc client and init v2ray api connection
	for {
		v2Client.HsClient, v2Client.SsClient, v2Client.CmdConn, err = v2rayApi.V2InitApi(baseCfg)
		if err != nil {
			log.Error(err)
		} else {
			break
		}
	}

	for {
		usersNow, err = GetUserSelector(idIndex)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		useRemove, userAdd, err := structures.FindUserDiffer(usersBefore, usersNow)
		if err != nil {
			log.Error(err)
		}

		if userAdd != nil {
			log.Debugf(fmt.Sprint("Add users ", *userAdd))
			err = v2Client.v2rayAddUsers(userAdd)
			if err != nil {
				log.Error(err)
			}
		}
		if useRemove != nil {
			log.Debugf(fmt.Sprint("Remove users ", *useRemove))
			err = v2Client.v2rayRemoveUsers(useRemove)
			if err != nil {
				log.Error(err)
			}
		}

		// Sync_interval
		time.Sleep(time.Duration(baseCfg.Sync.Interval) * time.Second)

		usersTraffic, err = v2Client.v2rayQueryTraffic(usersNow)
		if err != nil {
			log.Error(err)
		}
		_, err = PostUserSelector(idIndex, usersTraffic)
		if err != nil {
			log.Error(err)
		}
		usersBefore = usersNow
	}
}

func main() {
	log.Debug("start")

	for idIndex := 0; idIndex < len(baseCfg.Panel.NodeIDs); idIndex++ {
		go sync(uint32(idIndex))
	}

	//time.Sleep(time.Duration(100000) * time.Second)
}