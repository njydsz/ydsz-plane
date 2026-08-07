// Package dto — request DTOs for HTTP API (decouple handler params from domain models).
package dto

// --- Workspace ---

type CreateWorkspaceRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=80"`
	Slug     string `json:"slug" binding:"omitempty,max=60"`
	Timezone string `json:"timezone" binding:"omitempty"`
	Language string `json:"language" binding:"omitempty,oneof=zh-CN en-US en"`
}

type UpdateWorkspaceRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=80"`
	Timezone *string `json:"timezone,omitempty"`
	Language *string `json:"language,omitempty" binding:"omitempty,oneof=zh-CN en-US en"`
	LogoURL  *string `json:"logo_url,omitempty" binding:"omitempty,url,max=500"`
}

// --- Member ---

type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member guest"`
}

// --- Invitation ---

type SendInvitationRequest struct {
	Email   string `json:"email" binding:"required,email"`
	Role    string `json:"role" binding:"required,oneof=admin member guest"`
	Message string `json:"message" binding:"omitempty,max=500"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

// --- Project ---

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=80"`
	Slug        string `json:"slug" binding:"omitempty,max=60"`
	Identifier  string `json:"identifier" binding:"omitempty,max=6"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Network     string `json:"network" binding:"omitempty,oneof=public private"`
	Icon        string `json:"icon" binding:"omitempty,max=32"`
	Color       string `json:"color" binding:"omitempty,hexcolor,len=7"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=80"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Network     *string `json:"network,omitempty" binding:"omitempty,oneof=public private"`
	Icon        *string `json:"icon,omitempty"`
	Color       *string `json:"color,omitempty" binding:"omitempty,hexcolor,len=7"`
}
