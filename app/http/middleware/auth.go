package middleware

import (
	"github.com/ArtisanCloud/PowerX/app/http"
	"github.com/ArtisanCloud/PowerX/app/models"
	service "github.com/ArtisanCloud/PowerX/app/service"
	"github.com/ArtisanCloud/PowerX/app/service/wx/wecom"
	"github.com/ArtisanCloud/PowerX/config"
	"github.com/ArtisanCloud/PowerX/database"
	"github.com/gin-gonic/gin"
)

func AuthCustomerAPI(c *gin.Context) {

	apiResponse := http.NewAPIResponse(c)

	strAuthorization := c.GetHeader("Authorization")

	if strAuthorization == "" {
		apiResponse.SetCode(config.API_ERR_CODE_TOKEN_NOT_IN_HEADER, config.API_RETURN_CODE_ERROR, "", "")

	} else {
		var (
			customer *models.Customer
			err      error
		)
		ptrClaims, err := service.ParseAuthorization(strAuthorization)
		if ptrClaims == nil || err != nil {
			apiResponse.SetCode(config.API_ERR_CODE_ACCOUNT_INVALID_TOKEN, config.API_RETURN_CODE_ERROR, "", "")
			apiResponse.ThrowJSONResponse(c)
			return
		}
		claims := *ptrClaims
		if claims["OpenID"] == nil && claims["ExternalUserID"] == nil {
			apiResponse.SetCode(config.API_ERR_CODE_ACCOUNT_INVALID_TOKEN, config.API_RETURN_CODE_ERROR, "", "")
		} else {
			serviceWeComCustomer := wecom.NewWeComCustomerService(c)
			if claims["OpenID"] != nil {
				openID := claims["OpenID"].(string)
				if openID == "" {
					apiResponse.SetCode(config.API_ERR_CODE_LACK_OF_WX_EXTERNAL_USER_ID, config.API_RETURN_CODE_ERROR, "", "")
				}
				customer, err = serviceWeComCustomer.GetCustomerByOpenID(database.DBConnection, openID)

				// set auth open id
				wecom.SetAuthOpenID(c, openID)

			} else if claims["ExternalUserID"] != nil {
				externalUserID := claims["ExternalUserID"].(string)
				if externalUserID == "" {
					apiResponse.SetCode(config.API_ERR_CODE_LACK_OF_WX_EXTERNAL_USER_ID, config.API_RETURN_CODE_ERROR, "", "")
				}
				customer, err = serviceWeComCustomer.GetCustomerByWXExternalUserID(database.DBConnection, externalUserID)
			}

			if err != nil || customer.PowerModel == nil {
				apiResponse.SetCode(config.API_ERR_CODE_ACCOUNT_UNREGISTER, config.API_RETURN_CODE_ERROR, "", "")
			} else {
				service.SetAuthCustomer(c, customer)
			}

		}
	}

	if !apiResponse.IsNoError() {
		apiResponse.ThrowJSONResponse(c)
	}
	return

}

func AuthEmployeeAPI(c *gin.Context) {

	apiResponse := http.NewAPIResponse(c)
	strAuthorization := c.GetHeader("Authorization")
	if strAuthorization == "" {
		apiResponse.SetCode(config.API_ERR_CODE_TOKEN_NOT_IN_HEADER, config.API_RETURN_CODE_ERROR, "", "")

	} else {
		var (
			employee *models.Employee
			err      error
		)
		ptrClaims, err := service.ParseAuthorization(strAuthorization)
		if ptrClaims == nil || err != nil {
			apiResponse.SetCode(config.API_ERR_CODE_ACCOUNT_INVALID_TOKEN, config.API_RETURN_CODE_ERROR, "", "")
			apiResponse.ThrowJSONResponse(c)
			return
		}
		claims := *ptrClaims
		//fmt.Dump(claims)
		if claims["WXUserID"] == nil {
			apiResponse.SetCode(config.API_ERR_CODE_LACK_OF_WX_USER_ID, config.API_RETURN_CODE_ERROR, "", "")
			apiResponse.ThrowJSONResponse(c)
			return
		}
		wxUserID := claims["WXUserID"].(string)

		serviceWeComEmployee := wecom.NewWeComEmployeeService(c)

		if err != nil || wxUserID == "" {
			apiResponse.SetCode(config.API_ERR_CODE_LACK_OF_WX_USER_ID, config.API_RETURN_CODE_ERROR, "", "")
		} else {
			employee, err = serviceWeComEmployee.GetEmployeeByUserID(database.DBConnection, wxUserID)
			if err != nil || employee == nil {
				apiResponse.SetCode(config.API_ERR_CODE_EMPLOYEE_UNREGISTER, config.API_RETURN_CODE_ERROR, "", "")
			} else {
				service.SetAuthEmployee(c, employee)
			}
		}
	}

	if !apiResponse.IsNoError() {
		apiResponse.ThrowJSONResponse(c)
	}
	return

}

func AuthAdminAPI(c *gin.Context) {

}

// --------------------------------------------------------------------------------------------
func AuthWeb(c *gin.Context) {
	//resp := map[string]string{"hello":"world web"}
	//c.JSON(200, resp)

}