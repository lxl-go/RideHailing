package carpool

import (
	"errors"
	"strconv"

	carpoolReq "ride-hailing/admin-server/model/carpool/request"
	carpoolResp "ride-hailing/admin-server/model/carpool/response"
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/service"
	"ride-hailing/admin-server/utils"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewApi struct{}

var reviewService = service.ServiceGroupApp.CarpoolServiceGroup.ReviewService

func (a *ReviewApi) ListAudits(c *gin.Context) {
	var search carpoolReq.CertReviewListSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := reviewService.ListAudits(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("查询审核列表失败")
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(response.PageResult{
		List:     carpoolResp.NewCertificationAuditResponses(list),
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, c)
}

func (a *ReviewApi) GetAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("ID参数错误", c)
		return
	}
	audit, err := reviewService.GetAudit(c.Request.Context(), id)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("查询审核详情失败")
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(carpoolResp.NewCertificationAuditResponse(*audit), c)
}

func (a *ReviewApi) ApproveAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("ID参数错误", c)
		return
	}
	reviewerID := int64(utils.GetUserID(c))
	if err := reviewService.ApproveAudit(c.Request.Context(), id, reviewerID); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("审核通过失败")
		response.FailWithMessage(auditOperationFailureMessage(err), c)
		return
	}
	response.OkWithMessage("审核通过", c)
}

func (a *ReviewApi) RejectAudit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("ID参数错误", c)
		return
	}
	var act carpoolReq.CertReviewAction
	if err := c.ShouldBindJSON(&act); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if act.RejectReason == "" {
		response.FailWithMessage("驳回原因必填", c)
		return
	}
	reviewerID := int64(utils.GetUserID(c))
	if err := reviewService.RejectAudit(c.Request.Context(), id, reviewerID, act.RejectReason); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("审核驳回失败")
		response.FailWithMessage(auditOperationFailureMessage(err), c)
		return
	}
	response.OkWithMessage("已驳回", c)
}

func auditOperationFailureMessage(err error) string {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "审核记录不存在或已被处理，请刷新列表后重试"
	}
	return "操作失败，请稍后重试"
}

func (a *ReviewApi) ListVehicleReviews(c *gin.Context) {
	var search carpoolReq.VehicleReviewListSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := utils.Verify(search.PageInfo, utils.PageInfoVerify); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := reviewService.ListVehicleReviews(c.Request.Context(), search)
	if err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("查询车辆审核列表失败")
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(response.PageResult{
		List:     carpoolResp.NewVehicleInfoResponses(list),
		Total:    total,
		Page:     search.Page,
		PageSize: search.PageSize,
	}, c)
}

func (a *ReviewApi) HandleVehicleReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("ID鍙傛暟閿欒", c)
		return
	}
	var act carpoolReq.VehicleReviewAction
	if err := c.ShouldBindJSON(&act); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if act.Status == 2 && act.RejectReason == "" {
		response.FailWithMessage("驳回原因必填", c)
		return
	}
	reviewerID := int64(utils.GetUserID(c))
	if err := reviewService.HandleVehicleReview(c.Request.Context(), id, reviewerID, act.Status, act.RejectReason); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("carpool").Err(err).Error("车辆审核操作失败")
		response.FailWithMessage(auditOperationFailureMessage(err), c)
		return
	}
	response.OkWithMessage("操作成功", c)
}
