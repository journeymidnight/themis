package worker

import (
	"github.com/ljjjustin/themis/config"
	"github.com/ljjjustin/themis/database"
	"github.com/ljjjustin/themis/utils"

	"github.com/astaxie/beego/httplib"
	"errors"
)

var convergeDecisionMatrix = []bool{
	/* +------------+----------|--------------+ */
	/* | Management | Storage  |    Fence     | */
	/* +------------+----------+--------------+ */
	/* | good       | good     | */ false, /* | */
	/* | good       | bad      | */ true, /*  | */
	/* | bad        | good     | */ false, /* | */
	/* | bad        | bad      | */ true, /*  | */
	/* +-----------------------+--------------+ */
}

type ConvergeWorker struct {
	config         *config.ThemisConfig
	decisionMatrix []bool
	flagTagMap     map[string]uint
}

func NewConvergekWorker(config *config.ThemisConfig) *ConvergeWorker {

	var flagManage uint = 1 << 1
	var flagStorage uint = 1 << 0

	flagTagMap := map[string]uint{
		"manage":  flagManage,
		"storage": flagStorage,
	}

	return &ConvergeWorker{
		config:         config,
		decisionMatrix: convergeDecisionMatrix,
		flagTagMap:     flagTagMap,
	}
}

func (w *ConvergeWorker) GetDecision(host *database.Host, states []*database.HostState) bool {

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

func (w *ConvergeWorker) FenceHost(host *database.Host) {
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

	// update host status
	host.Status = utils.HostEvacuatingStatus
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

func (w *ConvergeWorker) Evacuate(host *database.Host) error {

	// evacuate all virtual machine on that host
	// send host name to catkeeper
	url := w.config.CatKeeper.Url + "/catkeeper/v1/servers/evacuate"
	req := httplib.Post(url)

	reqBody := struct {
		HostName string `json:"hostname"`
		User     string `json:"user"`
	}{
		host.Name,
		w.config.CatKeeper.Username,
	}

	req.JSONBody(reqBody)

	resp, err := req.DoRequest()
	if err != nil {
		plog.Warningf("send host : %s to catkeeper failed", host.Name)
		return err
	}

	if resp.StatusCode != 202 {
		plog.Warningf("catkeeper evacuate host : %s vm failed", host.Name)
		return errors.New("catkeeper evacuate host : " + host.Name + " vm failed")
	}

	return nil
}
