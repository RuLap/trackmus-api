package user

type GetUserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type GetUploadURLResponse struct {
	URL string
}

type ConfirmUploadAvatarResponse struct {
	URL string
}

type SaveUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string `json:"last_name" validate:"required,min=1,max=50"`
	Username  string `json:"username" validate:"required,min=1,max=50"`
}
