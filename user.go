package courses

type User struct {
	Id          int    `json:"-" db:"id"`
	LastName    string `json:"last_name" binding:"required" db:"last_name"`
	FirstName   string `json:"first_name" binding:"required" db:"first_name"`
	Email       string `json:"email" binding:"required" db:"email"`
	PhoneNumber string `json:"phone_number" binding:"required" db:"phone_number"`
	Password    string `json:"password" db:"password_hash"`
	Salt        int    `json:"-" db:"salt"`
	Vk          bool   `json:"vk" db:"vk"`
	VkId        int    `json:"vk_id" db:"vk_id"`
	Admin       bool   `json:"-" db:"is_admin"`
}
