package courses

type Administrator struct {
	Login    string `json:"admin_login" binding:"required"`
	Password string `json:"admin_password" binding:"required"`
}
