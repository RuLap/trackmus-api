package user

func UserToGetResponse(model *User, avatarURL string) GetUserResponse {
	return GetUserResponse{
		ID:        model.ID.String(),
		FirstName: model.FirstName,
		LastName:  model.LastName,
		Username:  model.Username,
		AvatarURL: avatarURL,
	}
}

func SaveRequestToUser(req *SaveUserRequest) User {
	return User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
	}
}
