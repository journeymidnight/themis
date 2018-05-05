package monitor

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
	"github.com/ljjjustin/themis/fence"
)

var doFenceStatus = HostFailedStatus

type WorkerInterface interface {
	GetDecision(host *database.Host, states []*database.HostState) bool
	FenceHost(host *database.Host, states []*database.HostState)
}

func NewWorker(config *config.ThemisConfig) WorkerInterface {

	return NewOpenstackWorker(config)
}

func powerOffHost(host *database.Host) error {

	// execute power off through IPMI
	fencers, err := database.FencerGetByHost(host.Id)
	if err != nil || len(fencers) < 1 {
		plog.Warning("Can't find fencers with given host: ", host.Name)
		return err
	}

	var IPMIFencers []fence.FencerInterface
	for _, fencer := range fencers {
		IPMIFencers = append(IPMIFencers, fence.NewFencer(fencer))
	}

	fenced := false
	plog.Debug("Begin execute fence operation")
	for _, fencer := range IPMIFencers {
		if err = fencer.Fence(); err != nil {
			plog.Warningf("Fence operation failed on host %s", host.Name)
			continue
		}
		fenced = true
		plog.Infof("Fence operation successed on host: %s", host.Name)
		break
	}
	if fenced {
		return nil
	} else {
		return err
	}
}
