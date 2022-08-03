package admin

import (
	database2 "github.com/ArtisanCloud/PowerLibs/v2/database"
	models2 "github.com/ArtisanCloud/PowerSocialite/v2/src/models"
	"github.com/ArtisanCloud/PowerX/app/http/controllers/api"
	"github.com/ArtisanCloud/PowerX/app/http/request"
	"github.com/ArtisanCloud/PowerX/app/models"
	"github.com/ArtisanCloud/PowerX/app/service"
	"github.com/ArtisanCloud/PowerX/app/service/wx/wecom"
	"github.com/ArtisanCloud/PowerX/config"
	"github.com/ArtisanCloud/PowerX/database"
	"github.com/gin-gonic/gin"
)

type EmployeeAPIController struct {
	*api.APIController
	ServiceEmployee *service.EmployeeService
}

func NewEmployeeAPIController(context *gin.Context) (ctl *EmployeeAPIController) {

	return &EmployeeAPIController{
		APIController:   api.NewAPIController(context),
		ServiceEmployee: service.NewEmployeeService(context),
	}
}

func APIGetEmployeeList(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	params, _ := context.Get("params")
	para := params.(request.ParaList)

	defer api.RecoverResponse(context, "api.admin.employee.list")

	arrayList, err := ctl.ServiceEmployee.GetList(database.DBConnection, nil, para.Page, para.PageSize)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_GET_EMPLOYEE_LIST, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, arrayList)
}

func APIGetEmployeeDetail(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	userIDInterface, _ := context.Get("userID")
	userID := userIDInterface.(string)

	defer api.RecoverResponse(context, "api.admin.employee.detail")

	employee, err := ctl.ServiceEmployee.GetEmployeeByUserID(database.DBConnection, userID)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_GET_EMPLOYEE_DETAIL, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, employee)
}

func APIUpsertEmployee(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	params, _ := context.Get("employee")
	employee := params.(*models.Employee)

	defer api.RecoverResponse(context, "api.admin.employee.upsert")

	var err error
	employee, err = ctl.ServiceEmployee.UpsertEmployee(database.DBConnection, employee)

	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_UPSERT_EMPLOYEE, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, employee)

}

func APIDeleteEmployees(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	uuids, _ := context.Get("uuids")

	defer api.RecoverResponse(context, "api.admin.employee.delete")

	employees, err := ctl.ServiceEmployee.GetEmployees(database.DBConnection, uuids.([]string))
	if len(employees) <= 0 {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_GET_EMPLOYEE_LIST, config.API_RETURN_CODE_ERROR, "", "")
		panic(ctl.RS)
		return
	}

	err = ctl.ServiceEmployee.DeleteEmployees(database.DBConnection, employees)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_DELETE_EMPLOYEE, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, err)
}

func APIBindCustomerToEmployee(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	customerInterface, _ := context.Get("customer")
	customer := customerInterface.(*models.Customer)
	employeeInterface, _ := context.Get("employee")
	employee := employeeInterface.(*models.Employee)
	followInfoInterface, _ := context.Get("followInfo")
	followInfo := followInfoInterface.(*models2.FollowUser)

	defer api.RecoverResponse(context, "api.admin.employee.bind.customer")

	pivot, err := ctl.ServiceEmployee.BindCustomerToEmployee(customer.ExternalUserID.String, followInfo)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_BIND_CUSOTMER_TO_EMPLOYEE, config.API_RETURN_CODE_ERROR, "", "")
		panic(ctl.RS)
		return
	}
	// save operation log
	_ = (&database2.PowerOperationLog{}).SaveOps(database.DBConnection, customer.Name, customer,
		service.MODULE_CUSTOMER, "系统绑定外部联系人与员工", database2.OPERATION_EVENT_CREATE,
		employee.Name, employee, database2.OPERATION_RESULT_SUCCESS)

	if len(followInfo.Tags) > 0 {
		serviceWXTag := wecom.NewWXTagService(nil)
		err = serviceWXTag.SyncWXTagsByFollowInfos(database.DBConnection, pivot, followInfo)
		if err != nil {
			ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_SYNC_WX_TAG, config.API_RETURN_CODE_ERROR, "", "")
			panic(ctl.RS)
			return
		}
	}

	ctl.RS.Success(context, err)
}

func APIUnbindCustomerToEmployee(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	customerInterface, _ := context.Get("customer")
	customer := customerInterface.(*models.Customer)
	employeeInterface, _ := context.Get("employee")
	employee := employeeInterface.(*models.Employee)

	defer api.RecoverResponse(context, "api.admin.employee.bind.customer")

	_, _, err := ctl.ServiceEmployee.UnbindCustomerToEmployee(customer.ExternalUserID.String, employee.WXUserID.String)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_BIND_CUSOTMER_TO_EMPLOYEE, config.API_RETURN_CODE_ERROR, "", "")
		panic(ctl.RS)
		return
	}
	// save operation log
	_ = (&database2.PowerOperationLog{}).SaveOps(database.DBConnection, customer.Name, customer,
		service.MODULE_CUSTOMER, "系统解绑外部联系人与员工", database2.OPERATION_EVENT_DELETE,
		employee.Name, employee, database2.OPERATION_RESULT_SUCCESS)

	ctl.RS.Success(context, err)
}

// ------------------------------------------------------------

func APIGetEmployeeListOnWXPlatform(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	departmentIDInterface, _ := context.Get("departmentID")
	departmentID := departmentIDInterface.(int)

	defer api.RecoverResponse(context, "api.admin.employee.list")

	arrayList, err := wecom.WeComEmployee.App.User.GetDepartmentUsers(departmentID, 1)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_GET_EMPLOYEE_LIST, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, arrayList)
}

func APIGetEmployeeDetailOnWXPlatform(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	userIDInterface, _ := context.Get("userID")
	userID := userIDInterface.(string)

	defer api.RecoverResponse(context, "api.admin.employee.detail")

	result, err := wecom.WeComEmployee.App.User.Get(userID)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_GET_EMPLOYEE_DETAIL, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, result)
}

func APIDeleteEmployeesOnWXPlatform(context *gin.Context) {
	ctl := NewEmployeeAPIController(context)

	userIDInterface, _ := context.Get("userID")
	userID := userIDInterface.(string)

	defer api.RecoverResponse(context, "api.admin.employee.delete")

	result, err := wecom.WeComEmployee.App.User.Delete(userID)
	if err != nil {
		ctl.RS.SetCode(config.API_ERR_CODE_FAIL_TO_DELETE_EMPLOYEE, config.API_RETURN_CODE_ERROR, "", err.Error())
		panic(ctl.RS)
		return
	}

	ctl.RS.Success(context, result)
}