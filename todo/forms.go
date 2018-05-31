package todo

type Credentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CredentialsConfirm struct {
	Credentials
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Credentials.Password"`
}
