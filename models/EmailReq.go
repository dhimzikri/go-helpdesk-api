package models

type EmailRequest struct {
	IsSendEmail  int    `json:"issendemail" binding:"required"`
	Status       string `json:"status" binding:"required"`
	CustomerName string `json:"customername" gorm:"column:customername"`
	TicketNo     string `json:"ticketno" binding:"required"`
	TranCodeID   string `json:"trancodeid" binding:"required"`
	Email        string `json:"email_" binding:"required"`
	UserID       string `json:"userid" binding:"required"`
}

type EmailData struct {
	Subject string `gorm:"column:subject_"` // Match the output column names
	Body    string `gorm:"column:body_"`
	Email   string `gorm:"column:email"` // Match the output column names
}

// type EmailRequest struct {
// 	IsSendEmail  int    `json:"issendemail" binding:"required"`
// 	Status       string `json:"status" binding:"required"`
// 	CustomerName string `json:"customername" gorm:"column:customername"`
// 	TicketNo     string `json:"ticketno" binding:"required"`
// 	TranCodeID   string `json:"trancodeid" binding:"required"`
// 	Email        string `json:"email_" binding:"required"`
// 	UserID       string `json:"userid" binding:"required"`
// }
