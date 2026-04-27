package domain

type Role string

const (
	RoleSuperAdmin    Role = "super_admin"
	RoleBusinessAdmin Role = "business_admin"
	RoleManager       Role = "manager"
	RoleEmployee      Role = "employee"
)
