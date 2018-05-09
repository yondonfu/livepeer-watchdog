package services

import (
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/glog"
	"github.com/yondonfu/livepeer-watchdog/contracts"
)

type RatesWatcher struct {
	bondingManager *contracts.BondingManager
	notifier       *TwilioService
	db             *DB
}

func NewRatesWatcher(bondingManager *contracts.BondingManager, notifier *TwilioService, db *DB) *RatesWatcher {
	return &RatesWatcher{
		bondingManager: bondingManager,
		notifier:       notifier,
		db:             db,
	}
}

func (rw *RatesWatcher) Watch() (ethereum.Subscription, error) {
	ch := make(chan *contracts.BondingManagerTranscoderUpdate)
	sub, err := rw.bondingManager.WatchTranscoderUpdate(nil, ch, []common.Address{})
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

				glog.Infof("Received a transcoder update event in block %v!", e.Raw.BlockNumber)

				watchers, err := rw.db.AllWatchers()
				if err != nil {
					glog.Error(err)
					return
				}

				for telNo, tAddr := range watchers {
					if e.Transcoder == tAddr {
						msg := fmt.Sprintf(
							"Transcoder %v updated its rates\nReward Cut: %v\nFee Share: %v\nPrice: %v",
							common.ToHex(tAddr[:]),
							e.PendingRewardCut,
							e.PendingFeeShare,
							e.PendingPricePerSegment,
						)
						err := rw.notifier.Notify(telNo, msg)
						if err != nil {
							glog.Errorf("Error with Twilio notification: %v", err)
						} else {
							glog.Infof("Notified %v of a transcoder update event in block %v", telNo, e.Raw.BlockNumber)
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
