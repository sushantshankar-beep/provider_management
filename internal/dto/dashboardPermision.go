package dto

type PermissionItemDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type PermissionResponse struct {
	ID          string              `json:"_id"`
	Category    string              `json:"category"`
	Label       string              `json:"label"`
	Permissions []PermissionItemDTO `json:"permissions"`
	Order       int                 `json:"order"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
}

type CreatePermissionRequest struct {
	Category    string              `json:"category" binding:"required"`
	Label       string              `json:"label" binding:"required"`
	Permissions []PermissionItemDTO `json:"permissions" binding:"required"`
	Order       int                 `json:"order"`
}

type UpdatePermissionRequest struct {
	Category    string              `json:"category"`
	Label       string              `json:"label"`
	Permissions []PermissionItemDTO `json:"permissions"`
	Order       int                 `json:"order"`
}