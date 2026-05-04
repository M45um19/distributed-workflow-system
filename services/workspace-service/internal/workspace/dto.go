package workspace

type CreateWorkspaceInput struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Slug        string `json:"slug" validate:"required,min=3"`
	Description string `json:"description" validate:"max=255"`
}
