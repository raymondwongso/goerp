package domain

type RolePermission struct {
	RoleID       string `json:"role_id"       db:"role_id"`
	PermissionID string `json:"permission_id" db:"permission_id"`
}
