package auth

func LoginRequestToUser(dto *LoginRequest, hashedPassword string) *User {
	return &User{
		Email:    dto.Email,
		Password: &hashedPassword,
	}
}

func RegisterRequestToUser(dto *RegisterRequest, hashedPassword string) *User {
	return &User{
		Email:          dto.Email,
		Password:       &hashedPassword,
		Provider:       LocalProvider,
		EmailConfirmed: false,
	}
}
