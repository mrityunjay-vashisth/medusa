package adminsvc

type DepartmentServicesInterface interface {
	CreateDepartment()
	UpdateDepartment()
	UpdateDepartmentById()
	DeleteDepartmentById()
	GetDepartments()
	GetDepartmentById()
}

func (a *adminService) CreateDepartment() {

}
