package carpool

import (
	"strconv"

	"github.com/gin-gonic/gin"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"
)

type PersonApi struct{}

var personService = service.ServiceGroupApp.CarpoolServiceGroup.PersonService

func (a *PersonApi) ListPersons(c *gin.Context) {
	var search carpoolReq.PersonSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := personService.ListPersons(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("list persons failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *PersonApi) GetPerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	detail, err := personService.GetPersonDetail(c.Request.Context(), id)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("get person failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(detail, c)
}

func (a *PersonApi) GetDriverStats(c *gin.Context) {
	data, err := personService.GetStats(c.Request.Context(), "driver")
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("get driver stats failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PersonApi) GetPassengerStats(c *gin.Context) {
	data, err := personService.GetStats(c.Request.Context(), "passenger")
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("get passenger stats failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *PersonApi) CreatePerson(c *gin.Context) {
	var payload carpoolReq.PersonPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	person, err := personService.CreatePerson(c.Request.Context(), payload)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("create person failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(person, c)
}

func (a *PersonApi) UpdatePerson(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("invalid id", c)
		return
	}
	var payload carpoolReq.PersonPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	person, err := personService.UpdatePerson(c.Request.Context(), id, payload)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("update person failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(person, c)
}

func (a *PersonApi) AssignRoles(c *gin.Context) {
	var req carpoolReq.PersonRoleAssign
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	person, err := personService.AssignRoles(c.Request.Context(), req)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("assign roles failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(person, c)
}

func (a *PersonApi) BatchStatus(c *gin.Context) {
	var req carpoolReq.PersonBatchStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > 100 {
		response.FailWithMessage("ids size must be 1-100", c)
		return
	}
	if err := personService.BatchUpdateStatus(c.Request.Context(), req); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("batch status failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (a *PersonApi) BatchDeleteDrivers(c *gin.Context) {
	var req carpoolReq.PersonBatchDelete
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if len(req.IDs) == 0 || len(req.IDs) > 100 {
		response.FailWithMessage("ids size must be 1-100", c)
		return
	}
	if err := personService.BatchDeleteDrivers(c.Request.Context(), req); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("batch delete drivers failed")
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.Ok(c)
}

func (a *PersonApi) PreviewImport(c *gin.Context) {
	var payload carpoolReq.PersonImportPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	result, err := personService.PreviewImport(c.Request.Context(), payload)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PersonApi) CommitImport(c *gin.Context) {
	var payload carpoolReq.PersonImportPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	result, err := personService.CommitImport(c.Request.Context(), payload)
	if err != nil {
		response.FailWithDetailed(result, err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *PersonApi) ListImportErrors(c *gin.Context) {
	var search carpoolReq.PersonImportErrorSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid params: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := personService.ListImportErrors(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("person").Err(err).Error("list import errors failed")
		response.FailWithMessage("query failed", c)
		return
	}
	response.OkWithData(response.PageResult{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, c)
}

func (a *PersonApi) ExportPersons(c *gin.Context) {
	response.OkWithData(gin.H{"taskId": personService.ExportPersons(c.Request.Context())}, c)
}
