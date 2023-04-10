package hardwarebundle

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaypipes/ghw"
	"github.com/sc-js/core_backend/src/tools"
)

var hwCache hardwareInformation
var hwExpire = time.Now()

func (con *hardwareController) getHardwareUsageHandler(c *gin.Context) {
	tools.GetPagedAndSend[hardwareUsage](c, con.DataWrap.DB)
}

func (con *hardwareController) getHardwareConfigurationHandler(c *gin.Context) {
	if runtime.GOOS == "darwin" {
		tools.RespondError(errors.New("not_supported_platform"), http.StatusNotImplemented, c)
		return
	}

	inv, err := strconv.ParseBool(c.Request.URL.Query().Get("invalidate"))
	if err != nil {
		inv = false
	}

	if hwExpire.Before(time.Now()) || inv {
		hwCache = hardwareInformation{}

		if data, err := ghw.CPU(); err == nil {
			hwCache.CPU = *data
		}
		if data, err := ghw.GPU(); err == nil {
			hwCache.GPU = *data
		}
		if data, err := ghw.Memory(); err == nil {
			hwCache.Memory = *data
		}
		if data, err := ghw.Block(); err == nil {
			hwCache.Storage = *data
		}
		if data, err := ghw.Network(); err == nil {
			hwCache.Network = *data
		}
		if data, err := ghw.Topology(); err == nil {
			hwCache.Topology = *data
		}
		if data, err := ghw.BIOS(); err == nil {
			hwCache.Bios = *data
		}
		if data, err := ghw.PCI(); err == nil {
			hwCache.PCI = *data
		}
		if data, err := ghw.Baseboard(); err == nil {
			hwCache.Baseboard = *data
		}
		if data, err := ghw.Chassis(); err == nil {
			hwCache.Chassis = *data
		}
		if data, err := ghw.Product(); err == nil {
			hwCache.Product = *data
		}

		hwExpire = time.Now().Add(time.Minute * 30)
	}

	tools.RespondWithJSON(c, http.StatusOK, hwCache)
}
