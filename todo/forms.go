package todo

// Credentials for login
type Credentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// CredentialsConfirm for signup
type CredentialsConfirm struct {
	Credentials
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Credentials.Password"`
}
