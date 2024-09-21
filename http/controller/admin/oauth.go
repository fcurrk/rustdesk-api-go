package admin

import (
	"Gwen/global"
	"Gwen/http/request/admin"
	adminReq "Gwen/http/request/admin"
	"Gwen/http/response"
	"Gwen/model"
	"Gwen/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type Oauth struct {
}

// Info
func (o *Oauth) Info(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		response.Fail(c, 101, "参数错误")
		return
	}
	v := service.AllService.OauthService.GetOauthCache(code)
	if v == nil {
		response.Fail(c, 101, "信息不存在")
		return
	}
	response.Success(c, v)
}

func (o *Oauth) ToBind(c *gin.Context) {
	f := &adminReq.BindOauthForm{}
	err := c.ShouldBindJSON(f)
	if err != nil {
		response.Fail(c, 101, "参数错误")
		return
	}
	u := service.AllService.UserService.CurUser(c)

	utr := service.AllService.UserService.UserThirdInfo(u.Id, f.Op)
	if utr.Id > 0 {
		response.Fail(c, 101, "已绑定过了")
		return
	}

	err, code, url := service.AllService.OauthService.BeginAuth(f.Op)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	service.AllService.OauthService.SetOauthCache(code, &service.OauthCacheItem{
		Action: service.OauthActionTypeBind,
		Op:     f.Op,
		UserId: u.Id,
	}, 5*60)

	response.Success(c, gin.H{
		"code": code,
		"url":  url,
	})
}

// Confirm 确认授权登录
func (o *Oauth) Confirm(c *gin.Context) {
	j := &adminReq.OauthConfirmForm{}
	err := c.ShouldBindJSON(j)
	if err != nil {
		response.Fail(c, 101, "参数错误"+err.Error())
		return
	}
	if j.Code == "" {
		response.Fail(c, 101, "参数错误: code 不存在")
		return
	}
	v := service.AllService.OauthService.GetOauthCache(j.Code)
	if v == nil {
		response.Fail(c, 101, "授权已过期")
		return
	}
	u := service.AllService.UserService.CurUser(c)
	v.UserId = u.Id
	service.AllService.OauthService.SetOauthCache(j.Code, v, 0)
	response.Success(c, v)
}

func (o *Oauth) BindConfirm(c *gin.Context) {
	j := &adminReq.OauthConfirmForm{}
	err := c.ShouldBindJSON(j)
	if err != nil {
		response.Fail(c, 101, "参数错误"+err.Error())
		return
	}
	if j.Code == "" {
		response.Fail(c, 101, "参数错误: code 不存在")
		return
	}
	v := service.AllService.OauthService.GetOauthCache(j.Code)
	if v == nil {
		response.Fail(c, 101, "授权已过期")
		return
	}
	u := service.AllService.UserService.CurUser(c)
	err = service.AllService.OauthService.BindGithubUser(v.ThirdOpenId, v.ThirdOpenId, u.Id)
	if err != nil {
		response.Fail(c, 101, "绑定失败，请重试")
		return
	}

	v.UserId = u.Id
	service.AllService.OauthService.SetOauthCache(j.Code, v, 0)
	response.Success(c, v)
}

func (o *Oauth) Unbind(c *gin.Context) {
	f := &adminReq.UnBindOauthForm{}
	err := c.ShouldBindJSON(f)
	if err != nil {
		response.Fail(c, 101, "参数错误")
		return
	}
	u := service.AllService.UserService.CurUser(c)
	utr := service.AllService.UserService.UserThirdInfo(u.Id, f.Op)
	if utr.Id == 0 {
		response.Fail(c, 101, "未绑定")
		return
	}
	if f.Op == model.OauthTypeGithub {
		err = service.AllService.OauthService.UnBindGithubUser(u.Id)
		if err != nil {
			response.Fail(c, 101, "解绑失败")
			return
		}
	}
	response.Success(c, nil)
}

// Detail Oauth
// @Tags Oauth
// @Summary Oauth详情
// @Description Oauth详情
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.Oauth}
// @Failure 500 {object} response.Response
// @Router /admin/oauth/detail/{id} [get]
// @Security token
func (o *Oauth) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	u := service.AllService.OauthService.InfoById(uint(iid))
	if u.Id > 0 {
		response.Success(c, u)
		return
	}
	response.Fail(c, 101, "信息不存在")
	return
}

// Create 创建Oauth
// @Tags Oauth
// @Summary 创建Oauth
// @Description 创建Oauth
// @Accept  json
// @Produce  json
// @Param body body admin.OauthForm true "Oauth信息"
// @Success 200 {object} response.Response{data=model.Oauth}
// @Failure 500 {object} response.Response
// @Router /admin/oauth/create [post]
// @Security token
func (o *Oauth) Create(c *gin.Context) {
	f := &admin.OauthForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, "参数错误"+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}

	ex := service.AllService.OauthService.InfoByOp(f.Op)
	if ex.Id > 0 {
		response.Fail(c, 101, "已存在"+f.Op)
		return
	}

	u := f.ToOauth()
	err := service.AllService.OauthService.Create(u)
	if err != nil {
		response.Fail(c, 101, "创建失败")
		return
	}
	response.Success(c, u)
}

// List 列表
// @Tags Oauth
// @Summary Oauth列表
// @Description Oauth列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} response.Response{data=model.OauthList}
// @Failure 500 {object} response.Response
// @Router /admin/oauth/list [get]
// @Security token
func (o *Oauth) List(c *gin.Context) {
	query := &admin.PageQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, "参数错误")
		return
	}
	res := service.AllService.OauthService.List(query.Page, query.PageSize, nil)
	response.Success(c, res)
}

// Update 编辑
// @Tags Oauth
// @Summary Oauth编辑
// @Description Oauth编辑
// @Accept  json
// @Produce  json
// @Param body body admin.OauthForm true "Oauth信息"
// @Success 200 {object} response.Response{data=model.OauthList}
// @Failure 500 {object} response.Response
// @Router /admin/oauth/update [post]
// @Security token
func (o *Oauth) Update(c *gin.Context) {
	f := &admin.OauthForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, "参数错误")
		return
	}
	if f.Id == 0 {
		response.Fail(c, 101, "参数错误")
		return
	}
	errList := global.Validator.ValidStruct(f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := f.ToOauth()
	err := service.AllService.OauthService.Update(u)
	if err != nil {
		response.Fail(c, 101, "更新失败")
		return
	}
	response.Success(c, nil)
}

// Delete 删除
// @Tags Oauth
// @Summary Oauth删除
// @Description Oauth删除
// @Accept  json
// @Produce  json
// @Param body body admin.OauthForm true "Oauth信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/oauth/delete [post]
// @Security token
func (o *Oauth) Delete(c *gin.Context) {
	f := &admin.OauthForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, "系统错误")
		return
	}
	id := f.Id
	errList := global.Validator.ValidVar(id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	u := service.AllService.OauthService.InfoById(f.Id)
	if u.Id > 0 {
		err := service.AllService.OauthService.Delete(u)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, err.Error())
		return
	}
	response.Fail(c, 101, "信息不存在")
}