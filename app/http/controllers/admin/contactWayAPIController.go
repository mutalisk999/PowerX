package admin

import (
	"github.com/ArtisanCloud/PowerX/app/http/controllers/api"
	"github.com/ArtisanCloud/PowerX/app/models"
	modelWX "github.com/ArtisanCloud/PowerX/app/models/wx"
	"github.com/ArtisanCloud/PowerX/app/service"
	"github.com/ArtisanCloud/PowerX/app/service/wx/weCom"
	global2 "github.com/ArtisanCloud/PowerX/config"
	"github.com/ArtisanCloud/PowerX/database/global"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon"
	"gorm.io/gorm/clause"
)

type ContactWayAPIController struct {
	*api.APIController
	ServiceContactWay *service.ContactWayService
	ServiceWXTag      *weCom.WXTagService
}

func NewContactWayAPIController(context *gin.Context) (ctl *ContactWayAPIController) {

	return &ContactWayAPIController{
		APIController:     api.NewAPIController(context),
		ServiceContactWay: service.NewContactWayService(context),
		ServiceWXTag:      weCom.NewWXTagService(context),
	}
}

func APIContactWaySync(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	defer api.RecoverResponse(context, "api.admin.wxTagGroup.sync")

	startDatetimeInterface, _ := context.Get("startDatetime")
	startDatetime := startDatetimeInterface.(*carbon.Carbon)
	endDatetimeInterface, _ := context.Get("endDatetime")
	endDatetime := endDatetimeInterface.(*carbon.Carbon)
	limitInterface, _ := context.Get("limit")
	limit := limitInterface.(int)

	err := ctl.ServiceContactWay.SyncContactWayFromWXPlatform(startDatetime, endDatetime, limit)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_SYNC_CONTACT_WAY_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, err)
}

func APIGetContactWayList(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	params, _ := context.Get("groupUUID")
	groupUUID := params.(string)

	defer api.RecoverResponse(context, "api.admin.contactWay.list")

	arrayList, err := ctl.ServiceContactWay.GetList(global.G_DBConnection, groupUUID)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_LIST, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, arrayList)
}

func APIGetContactWayDetail(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	configIDInterface, _ := context.Get("configID")
	configID := configIDInterface.(string)

	defer api.RecoverResponse(context, "api.admin.contactWay.detail")

	contactWay, err := ctl.ServiceContactWay.GetContactWayByConfigID(global.G_DBConnection, configID)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_DETAIL, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}
	if contactWay == nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_DETAIL, global2.API_RETURN_CODE_ERROR, "", "")
		panic(ctl.RS)
		return
	}

	contactWay.WXTags, err = contactWay.LoadWXTags(global.G_DBConnection, nil)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_DETAIL, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, contactWay)
}

func APICreateContactWay(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	params, _ := context.Get("contactWay")
	contactWay := params.(*models.ContactWay)

	defer api.RecoverResponse(context, "api.admin.contactWay.upsert")

	var err error

	// upload wechat contact way
	result, err := ctl.ServiceContactWay.CreateContactWayOnWXPlatform(contactWay)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_CREATE_CONTACT_WAY_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	contactWay, err = ctl.ServiceContactWay.ConvertResponseToContactWay(contactWay, result)
	//contactWay.ConfigID = "3495QhQRnTDdkOBwtBsmwLmNaC9plvlnQayZgb4k"

	// insert contact way
	err = ctl.ServiceContactWay.UpsertContactWays(global.G_DBConnection.Omit(clause.Associations), []*models.ContactWay{contactWay}, nil)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_CREATE_CONTACT_WAY, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}
	//err = ctl.ServiceWXTag.SyncWXTagsToObject(global.G_DBConnection, contactWay, contactWay.WXTags)
	err = ctl.ServiceWXTag.AppendWXTagsToObject(global.G_DBConnection, contactWay, contactWay.WXTags)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_SYNC_WX_TAG, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}
	ctl.RS.Success(context, contactWay)

}

func APIUpdateContactWay(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	contactWayInterface, _ := context.Get("contactWay")
	contactWay := contactWayInterface.(*models.ContactWay)

	updateTagsInterface, _ := context.Get("updateTags")
	updateTags := updateTagsInterface.([]*modelWX.WXTag)

	defer api.RecoverResponse(context, "api.admin.contactWay.upsert")

	var err error

	// upload wechat contact way
	err = ctl.ServiceContactWay.UpdateContactWayOnWXPlatform(contactWay)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_UPDATE_CONTACT_WAY_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	// update contact way
	err = ctl.ServiceContactWay.UpsertContactWays(global.G_DBConnection, []*models.ContactWay{contactWay}, nil)

	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_UPDATE_CONTACT_WAY, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	// replace wechat contact way tags
	err = ctl.ServiceWXTag.SyncWXTagsToObject(global.G_DBConnection, contactWay, updateTags)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_SYNC_WX_TAG_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, contactWay)

}

func APIDeleteContactWays(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	contactWayInterface, _ := context.Get("contactWay")
	contactWay := contactWayInterface.(*models.ContactWay)

	defer api.RecoverResponse(context, "api.admin.contactWay.delete")

	var err error
	// upload delete wechat contact way
	err = ctl.ServiceContactWay.DeleteContactWayOnWXPlatform(contactWay.ConfigID)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_DELETE_CONTACT_WAY_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	// clear contact way tags
	err = ctl.ServiceWXTag.ClearObjectWXTags(global.G_DBConnection, contactWay)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_DELETE_TAG, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	// delete contact way
	err = ctl.ServiceContactWay.DeleteContactWayByConfigID(global.G_DBConnection, contactWay.ConfigID)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_DELETE_CONTACT_WAY, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, err)
}

// ------------------------------------------------------------

func APIGetContactWayListOnWXPlatform(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	startDatetimeInterface, _ := context.Get("startDatetime")
	startDatetime := startDatetimeInterface.(*carbon.Carbon)
	endDatetimeInterface, _ := context.Get("endDatetime")
	endDatetime := endDatetimeInterface.(*carbon.Carbon)
	limitInterface, _ := context.Get("limit")
	limit := limitInterface.(int)

	defer api.RecoverResponse(context, "api.admin.contactWay.list")

	arrayList, err := ctl.ServiceContactWay.GetContactWayListOnWXPlatform(startDatetime, endDatetime, limit)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_LIST, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, arrayList)
}

func APIGetContactWayDetailOnWXPlatform(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	configIDInterface, _ := context.Get("configID")
	configID := configIDInterface.(string)

	responseContactWay, err := weCom.G_WeComApp.App.ExternalContactContactWay.Get(configID)

	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_GET_CONTACT_WAY_DETAIL, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, responseContactWay)

}

func APIDeleteContactWaysOnWXPlatform(context *gin.Context) {
	ctl := NewContactWayAPIController(context)

	configIDInterface, _ := context.Get("configID")
	configID := configIDInterface.(string)

	defer api.RecoverResponse(context, "api.admin.contactWay.delete")

	var err error
	// upload delete wechat contact way
	err = ctl.ServiceContactWay.DeleteContactWayOnWXPlatform(configID)
	if err != nil {
		ctl.RS.SetCode(global2.API_ERR_CODE_FAIL_TO_DELETE_CONTACT_WAY_ON_WX_PLATFORM, global2.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, err)
}
