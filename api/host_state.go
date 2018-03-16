package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ljjjustin/themis/storage"
)

func init() {
	Router().POST("/hosts/:id/states", CreateState)
	Router().GET("/hosts/:id/states", GetHostStates)
	Router().PUT("/hosts/:id/states/:sid", UpdateState)
	Router().DELETE("/hosts/:id/states/:sid", DeleteState)
}

func CreateState(c *gin.Context) {
	var state storage.HostState
	ParseBody(c, &state)

	host := GetHost(c)
	state.HostId = host.Id

	states, err := storage.StateGetAll(host.Id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}
	for _, s := range states {
		if state.Tag == s.Tag {
			AbortWithError(http.StatusBadRequest, ErrDuplicatedTag)
		}
	}

	if err := storage.StateInsert(&state); err != nil {
		AbortWithError(http.StatusNotAcceptable, err)
	} else {
		c.JSON(http.StatusOK, state)
	}
}

func GetHostStates(c *gin.Context) {
	host := GetHost(c)

	states, err := storage.StateGetAll(host.Id)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	}

	if tag := c.Query("tag"); len(tag) > 0 {
		filtered := make([]*storage.HostState, 0)
		for _, s := range states {
			if s.Tag == tag {
				filtered = append(filtered, s)
			}
		}
		c.JSON(http.StatusOK, filtered)
	} else {
		c.JSON(http.StatusOK, states)
	}
}

func UpdateState(c *gin.Context) {
	stateId := GetId(c, "sid")

	state, err := storage.StateGetById(stateId)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else if state == nil {
		AbortWithError(http.StatusNotFound, err)
	}

	ParseBody(c, state)
	err = storage.StateUpdate(stateId, state)
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(200, state)
	}
}

func DeleteState(c *gin.Context) {
	err := storage.StateDelete(GetId(c, "sid"))
	if err != nil {
		AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.Data(204, "application/json", make([]byte, 0))
	}
}
