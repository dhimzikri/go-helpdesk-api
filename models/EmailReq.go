package models

import "time"

// type EmailRequest struct {
// 	IsSendEmail  int    `json:"issendemail" binding:"required"`
// 	Status       string `json:"status" binding:"required"`
// 	CustomerName string `json:"customername" gorm:"column:customername"`
// 	TicketNo     string `json:"ticketno" binding:"required"`
// 	TranCodeID   string `json:"trancodeid" binding:"required"`
// 	Email        string `json:"email_" binding:"required"`
// 	UserID       string `json:"userid" binding:"required"`
// 	Subject      string `gorm:"column:subject_"`
// 	Body         string `gorm:"column:body_"`
// }

// type EmailData struct {
// 	Subject string `gorm:"column:subject_"` // Match the output column names
// 	Body    string `gorm:"column:body_"`
// 	Email   string `gorm:"column:email"` // Match the output column names
// }

type EmailRequest struct {
	IsSendEmail  int    `json:"issendemail" binding:"required"`
	Status       string `json:"status" binding:"required"`
	CustomerName string `json:"customername" gorm:"column:customername"`
	TicketNo     string `json:"ticketno" binding:"required"`
	TranCodeID   string `json:"trancodeid" binding:"required"`
	Email_       string `json:"email_" binding:"required"`
	UserID       string `json:"userid" binding:"required"`
	Subject      string `gorm:"column:subject_"` // Match the output column names
	Body         string `gorm:"column:body_"`
	Email        string `gorm:"column:email"` // Match the output column names

}

type EmailRecord struct {
	TicketNo string    `gorm:"column:ticketno"`
	Subject  string    `gorm:"column:subject_"`
	Body     string    `gorm:"column:body_"`
	EmailTo  string    `gorm:"column:emailto"`
	Flag     string    `gorm:"column:flag"`
	IsSend   int       `gorm:"column:issend"`
	UsrUpd   string    `gorm:"column:usrupd"`
	DtmUpd   time.Time `gorm:"column:dtmupd"`
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
