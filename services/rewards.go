package services

import (
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/yondonfu/livepeer-watchdog/contracts"
)

type RewardWatcher struct {
	bondingManager *contracts.BondingManager
	notifier       *TwilioService
	db             *DB
}

func NewRewardWatcher(bondingManager *contracts.BondingManager, notifier *TwilioService, db *DB) *RewardWatcher {
	return &RewardWatcher{
		bondingManager: bondingManager,
		notifier:       notifier,
		db:             db,
	}
}

func (rw *RewardWatcher) Watch() (ethereum.Subscription, error) {
	ch := make(chan *contracts.BondingManagerReward)
	sub, err := rw.bondingManager.WatchReward(nil, ch, []common.Address{})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case e, ok := <-ch:
				if !ok {
					glog.Error(err)
					return
				}

				glog.Infof("Received a reward event in block %v!", e.Raw.BlockNumber)

				watchers, err := rw.db.AllWatchers()
				if err != nil {
					glog.Error(err)
					return
				}

				for telNo, tAddr := range watchers {
					if e.Transcoder == tAddr {
						msg := fmt.Sprintf("Transcoder %v called reward in block %v", common.ToHex(tAddr[:]), e.Raw.BlockNumber)
						err := rw.notifier.Notify(telNo, msg)
						if err != nil {
							glog.Errorf("Error with Twilio notification: %v", err)
						} else {
							glog.Infof("Notified %v of a reward event in block %v", telNo, e.Raw.BlockNumber)
						}
					}
				}
			case <-sub.Err():
				return
			}
		}
	}()

	return sub, nil
}
