package service

import (
	"errors"
	modelPowerLib "github.com/ArtisanCloud/PowerLibs/v2/authorization/rbac/models"
	"github.com/ArtisanCloud/PowerLibs/v2/database"
	"github.com/ArtisanCloud/PowerX/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleService struct {
	*Service
	Role *modelPowerLib.Role
}

func NewRoleService(ctx *gin.Context) (r *RoleService) {
	r = &RoleService{
		Service: NewService(ctx),
		Role:    modelPowerLib.NewRole(nil),
	}
	return r
}

func (srv *RoleService) GetTreeList(db *gorm.DB, parentRoleID *string, needQueryChildren bool) (arrayRoles []*modelPowerLib.Role, err error) {

	arrayRoles = []*modelPowerLib.Role{}

	var conditions *map[string]interface{} = nil
	if parentRoleID != nil {
		conditions = &map[string]interface{}{
			"parent_id": parentRoleID,
		}
	}

	arrayRoles, err = srv.Role.GetTreeList(db, conditions, nil, modelPowerLib.ROLE_TYPE_ALL, parentRoleID, needQueryChildren)

	return arrayRoles, err
}

func (srv *RoleService) UpsertRoles(db *gorm.DB, roles []*modelPowerLib.Role, fieldsToUpdate []string) error {

	return database.UpsertModelsOnUniqueID(db, &modelPowerLib.Role{}, modelPowerLib.ROLE_UNIQUE_ID, roles, fieldsToUpdate)
}

func (srv *RoleService) DeleteRolesByIDs(db *gorm.DB, roleIDs []string) error {
	db = db.
		//Debug().
		Where("index_role_id in (?)", roleIDs).
		Where("type", modelPowerLib.ROLE_TYPE_NORMAL).
		Delete(&modelPowerLib.Role{})

	return db.Error
}

func (srv *RoleService) DeleteRoleByID(db *gorm.DB, roleID string) error {
	db = db.
		//Debug().
		Where("index_role_id", roleID).
		Delete(&modelPowerLib.Role{})

	return db.Error
}

func (srv *RoleService) GetRolesByIDs(db *gorm.DB, arrayRoleIDs []string) (roles []*modelPowerLib.Role, err error) {
	roles = []*modelPowerLib.Role{}

	if len(arrayRoleIDs) > 0 {
		db = db.
			//Debug().
			Where("index_role_id in (?)", arrayRoleIDs).
			Find(&roles)
		err = db.Error
	}

	return roles, err
}

func (srv *RoleService) GetRoleByID(db *gorm.DB, roleID string) (role *modelPowerLib.Role, err error) {
	role = &modelPowerLib.Role{}

	condition := &map[string]interface{}{
		"index_role_id": roleID,
	}
	err = database.GetFirst(db, condition, role, nil)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return role, err

}

func (srv *RoleService) GetEmployeeIDsByRoleIDs(db *gorm.DB, roleIDs []string) (employeeIDs []string, err error) {
	employeeIDs = []string{}
	result := db.Model(models.Employee{}).
		Debug().
		Where("role_id in (?)", roleIDs).
		Pluck("wx_user_id", &employeeIDs)

	return employeeIDs, result.Error
}

func (srv *RoleService) BindRoleToEmployeesByEmployeeIDs(db *gorm.DB, role *modelPowerLib.Role, employeeIDs []string) (err error) {

	result := db.Model(models.Employee{}).
		Debug().
		Where("wx_user_id in (?)", employeeIDs).
		Update("role_id", role.UniqueID)

	return result.Error
}
