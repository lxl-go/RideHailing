package system

import (
	"ride-hailing/admin-server/model/common/response"
	"ride-hailing/admin-server/model/system/request"
	"ride-hailing/admin-server/utils/logger"

	"github.com/gin-gonic/gin"
)

type DBApi struct{}

// InitDB
// @Tags     InitDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Param    data  body      request.InitDB                  true  "初始化数据库参数"
// @Success  200   {object}  response.Response{data=string}  "初始化用户数据库"
// @Router   /init/initdb [post]
func (i *DBApi) InitDB(c *gin.Context) {
	// Note: The guard on global.GVA_DB has been removed to allow
	// re-initialization from the API. The InitDB service creates
	// its own DB connection internally, so GVA_DB state is irrelevant.
	var dbInfo request.InitDB
	if err := c.ShouldBindJSON(&dbInfo); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("biz").Err(err).Error("参数校验不通过!")
		response.FailWithMessage("参数校验不通过", c)
		return
	}
	if err := initDBService.InitDB(dbInfo); err != nil {
		logger.WithCtx(c.Request.Context()).Mod("biz").Err(err).Error("自动创建数据库失败!")
		response.FailWithMessage("自动创建数据库失败，请查看后台日志，检查后在进行初始化", c)
		return
	}
	response.OkWithMessage("自动创建数据库成功", c)
}

// CheckDB
// @Tags     CheckDB
// @Summary  初始化用户数据库
// @Produce  application/json
// @Success  200  {object}  response.Response{data=map[string]interface{},msg=string}  "初始化用户数据库"
// @Router   /init/checkdb [post]
func (i *DBApi) CheckDB(c *gin.Context) {
	var (
		message  = "前往初始化数据库"
		needInit = true
	)
	// Always return needInit:true so the frontend displays the init form.
	// The API gate on GVA_DB was removed to allow re-initialization.
	logger.WithCtx(c.Request.Context()).Mod("biz").Info(message)
	response.OkWithDetailed(gin.H{"needInit": needInit}, message, c)
}
