package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang/glog"
	"github.com/yondonfu/livepeer-watchdog/contracts"
	"github.com/yondonfu/livepeer-watchdog/services"
)

const MainnetBondingManagerAddr = "0x511bc4556d823ae99630ae8de28b9b80df90ea2e"
const RinkebyBondingManagerAddr = "0xf6b0ceb5e3f25b6fbecf8186f8a68b4e42a96a17"

func loadConfig(fname string) (services.TwilioConfig, error) {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		return services.TwilioConfig{}, err
	}

	var config services.TwilioConfig
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return services.TwilioConfig{}, err
	}

	return config, nil
}

func watchdogMain(isRinkeby bool) error {
	config, err := loadConfig("config.json")
	if err != nil {
		return err
	}

	notifier := services.NewTwilioService(config)

	var bondingManagerAddr string

	if isRinkeby {
		glog.Infof("Connecting to Rinkeby Ethereum test network")

		bondingManagerAddr = RinkebyBondingManagerAddr
	} else {
		glog.Infof("Connecting to Ethereum main network")

		bondingManagerAddr = MainnetBondingManagerAddr
	}

	ethUrl := "ws://localhost:8546"

	backend, err := ethclient.Dial(ethUrl)
	if err != nil {
		return err
	}

	bondingManager, err := contracts.NewBondingManager(common.HexToAddress(bondingManagerAddr), backend)
	if err != nil {
		return err
	}

	db, err := services.NewDB("watchers.sqlite3")
	if err != nil {
		return err
	}

	server := services.NewWebServer("3000", db)
	server.Start()

	rewardWatcher := services.NewRewardWatcher(bondingManager, notifier, db)
	rewardSub, err := rewardWatcher.Watch()
	if err != nil {
		return err
	}

	ratesWatcher := services.NewRatesWatcher(bondingManager, notifier, db)
	ratesSub, err := ratesWatcher.Watch()
	if err != nil {
		return err
	}

	glog.Infof("Watching...")

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	select {
	case sig := <-ch:
		glog.Infof("Exiting: %v", sig)

		rewardSub.Unsubscribe()
		ratesSub.Unsubscribe()

		return nil
	}
}

func main() {
	flag.Set("logtostderr", "true")

	rinkeby := flag.Bool("rinkeby", false, "Set to true to connect to the Rinkeby Ethereum test network")

	flag.Parse()

	if err := watchdogMain(*rinkeby); err != nil {
		glog.Error(err)

		os.Exit(1)
	}
}
