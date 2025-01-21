package models

type User struct {
	ID        uint   `json:"id" gorm:"primaryKey;column:user_id"`
	Username  string `json:"user_name" gorm:"column:user_name"`
	GroupID   string `json:"group_id" gorm:"column:group_id"`
	RealName  string `json:"real_name" gorm:"column:real_name"`
	Password  string `json:"user_password" gorm:"column:user_password"`
	ConfinsID string `json:"confins_id" gorm:"column:confins_id"`
	IsActive  bool   `json:"is_active" gorm:"column:is_active"`
}

type UserLogin struct {
	User_name    string `json:"user_name"`
	UserPassword string `json:"user_password"`
	RealName     string `json:"realname"`
	Greetings    string `json:"greetings"`
	LastLogin    string `json:"last_login"`
	Email        string `json:"email"`
	Base64qrcode string `json:"base64qrcode"`
	HP           string `json:"hp"`
	Bucketcoll   string `json:"bucketcoll"`
	Token        string `json:"token"`
}

type Login struct {
	Username string `json:"username" example:"1879"`
	Password string `json:"pwd" example:"pass,123"`
}

type MyProfile struct {
	ID   string `json:"username"`
	Name string `json:"real_name"`
}
