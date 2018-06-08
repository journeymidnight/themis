package worker

import (
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/services"
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
	"github.com/ljjjustin/themis/utils"
)

var openstackDecisionMatrix = []bool{
	/* +------------+----------+-----------+--------------+ */
	/* | Management | Storage  | Network   |    Fence     | */
	/* +------------+----------+-----------+--------------+ */
	/* | good       | good     | good      | */ false, /* | */
	/* | good       | good     | bad       | */ true, /*  | */
	/* | good       | bad      | good      | */ true, /*  | */
	/* | good       | bad      | bad       | */ true, /*  | */
	/* | bad        | good     | good      | */ false, /* | */
	/* | bad        | good     | bad       | */ true, /*  | */
	/* | bad        | bad      | good      | */ true, /*  | */
	/* | bad        | bad      | bad       | */ true, /*  | */
	/* +-----------------------------------+--------------+ */
}

type OpenstackWorker struct {
	config         *config.ThemisConfig
	decisionMatrix []bool
	flagTagMap     map[string]uint
}

func NewOpenstackWorker(config *config.ThemisConfig) *OpenstackWorker {

	var flagManage uint = 1 << 2
	var flagStorage uint = 1 << 1
	var flagNetwork uint = 1 << 0

	flagTagMap := map[string]uint{
		"manage":  flagManage,
		"storage": flagStorage,
		"network": flagNetwork,
	}

	return &OpenstackWorker{
		config:         config,
		decisionMatrix: openstackDecisionMatrix,
		flagTagMap:     flagTagMap,
	}
}

func (w *OpenstackWorker) GetDecision(host *database.Host, states []*database.HostState) bool {

	if host.Disabled {
		return false
	}

	statusDecision := false
	if host.Status == doFenceStatus {
		statusDecision = true
	}

	var decision uint = 0
	for _, s := range states {
		// judge if one network is down.
		if s.FailedTimes >= 6 {
			decision |= w.flagTagMap[s.Tag]
		}
	}

	return statusDecision && w.decisionMatrix[decision]
}

func (w *OpenstackWorker) FenceHost(host *database.Host) {
	defer func() {
		if err := recover(); err != nil {
			plog.Warning("unexpected error during HandleEvents: ", err)
		}
	}()

	// check if we have disabled fence operation globally
	if w.config.Fence.DisableFenceOps {
		plog.Info("fence operation have been disabled.")
		return
	}

	plog.Infof("Begin fence host %s", host.Name)
	// update host status
	host.Status = utils.HostFencingStatus
	saveHost(host)

	err := powerOffHost(host)
	if err != nil {
		return
	}

	// update host status
	host.Status = utils.HostFencedStatus
	saveHost(host)

	err = w.Evacuate(host)
	if err != nil {
		return
	}

	// disable host status
	host.Status = utils.HostEvacuatedStatus
	host.Disabled = true
	saveHost(host)
}

func (w *OpenstackWorker) Evacuate(host *database.Host) error {

	// evacuate all virtual machine on that host
	nova, err := NewNovaClient(&w.config.Openstack)
	if err != nil {
		plog.Warning("Can't create nova client: ", err)
		return err
	}

	for {
		var computeService services.Service
		services, err := nova.ListServices()
		if err != nil {
			plog.Warning("Can't get service list", err)
			return err
		}
		for _, service := range services {
			if host.Name == service.Host && service.Binary == "nova-compute" {
				computeService = service
				break
			}
		}
		status := strings.ToLower(computeService.Status)
		if status != "disabled" {
			nova.DisableService(computeService, "disabled by themis monitor")
			plog.Infof("disable nova-compute on host :", host.Name)
		}

		state := strings.ToLower(computeService.State)
		if state != "up" {
			plog.Infof("nova-compute is down on host :", host.Name)
			break
		}

		time.Sleep(10 * time.Second)
	}

	// update host status
	host.Status = utils.HostEvacuatingStatus
	saveHost(host)

	servers, err := nova.ListServers(host.Name)
	if err != nil {
		plog.Warning("Can't get service list: ", err)
		return err
	}
	plog.Infof("get nova servers number :", len(servers))

	for _, server := range servers {
		id := server.ID
		plog.Infof("Try to evacuate instance: %s", id)
		nova.Evacuate(id)
	}

	return nil
}
