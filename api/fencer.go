package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/database"
	worker "github.com/ljjjustin/themis/worker"
	"github.com/ljjjustin/themis/config"
	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/ljjjustin/themis", "api")

func init() {
	Router().POST("/fencer/host/:id", FenceOneHost)
}

func FenceOneHost(c *gin.Context) {

	plog.Infof("fence host api begin.")

	id := GetId(c, "id")
	plog.Infof("get host id %d", id)

	var conf ConfigFile

	ParseBody(c, &conf)

	host, err := database.HostGetById(id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if host == nil {
		AbortWithError(http.StatusNotFound, err)
	}
	plog.Infof("get host %s success", host.Name)

	states, err := database.StateGetAll(id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if states == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	// load configurations
	themisCfg := config.NewConfig(conf.ConfigFile)

	//worker := monitor.NewWorker(themisCfg)
	worker := worker.NewWorker(themisCfg)

	if worker.GetDecision(host, states) {
		plog.Infof("into fencer host worker")
		worker.FenceHost(host)
		c.JSON(http.StatusAccepted, host)
	} else {
		c.JSON(http.StatusNotAcceptable, host)
	}
}

type ConfigFile struct {
	ConfigFile string `json:"config_file"`
}