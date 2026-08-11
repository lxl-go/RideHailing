package system

import (
	"github.com/gin-gonic/gin"

	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/utils/logger"
)

type GvaGovernanceApi struct{}

func (a *GvaGovernanceApi) GetGovernanceSummary(c *gin.Context) {
	data, err := gvaGovernanceService.GetGovernanceSummary(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get governance summary failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetRouteSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetRouteSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get route snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetAuditSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetAuditSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get audit snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetDatasourceSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetDatasourceSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get datasource snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) GetTimedTaskSnapshot(c *gin.Context) {
	data, err := gvaGovernanceService.GetTimedTaskSnapshot(c.Request.Context())
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("gvaGovernance").Err(err).Error("get timed task snapshot failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *GvaGovernanceApi) ExportGovernance(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": gvaGovernanceService.ExportGovernance(c.Request.Context())}, c)
}
