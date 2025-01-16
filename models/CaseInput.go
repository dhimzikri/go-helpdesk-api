package models

import "time"

type Case struct {
	AgreementNo   string `json:"AgreementNo" gorm:"column:agreementno"`
	ApplicationID string `json:"ApplicationID" gorm:"column:applicationid"`
	BranchID      string `json:"BranchID" binding:"required" gorm:"column:branchid"`
	CallerID      string `json:"CallerID" gorm:"column:callerid"`
	// ContactDescription  string    `json:"ContactDescription"`
	ContactID    int    `json:"ContactID" gorm:"column:applicationid"`
	CustomerID   string `json:"CustomerID" gorm:"column:customerid"`
	CustomerName string `json:"CustomerName" gorm:"column:customername"`
	DateCr       string `json:"DateCr" binding:"required" gorm:"column:date_cr"`
	Description  string `json:"Description" gorm:"column:description"`
	Email        string `json:"Email" gorm:"column:email"`
	Email_       string `json:"Email_" gorm:"column:email_"`
	FlagCompany  string `json:"FlagCompany" binding:"required" gorm:"column:flagcompany"`
	PhoneNo      string `json:"PhoneNo" gorm:"column:phoneno"`
	// PriorityDescription string    `json:"PriorityDescription"`
	PriorityID int `json:"PriorityID" gorm:"column:priorityid"`
	// RelationDescription string    `json:"RelationDescription"`
	RelationID   int    `json:"RelationID" gorm:"column:relationid"`
	RelationName string `json:"RelationName" gorm:"column:relationname"`
	// StatusDescription   *string   `json:"StatusDescription"`
	StatusID int `json:"StatusID" binding:"required" gorm:"column:statusid"`
	// StatusName          string    `json:"StatusName"`
	SubTypeID int    `json:"SubTypeID" gorm:"column:subtypeid"`
	TicketNo  string `json:"TicketNo" gorm:"primaryKey;column:ticketno"`
	// TypeDescription     string    `json:"TypeDescription"`
	TypeID int `json:"TypeID"  gorm:"column:typeid"`
	// TypeSubDescription  string    `json:"TypeSubDescription"`
	UserID string `json:"UserID" binding:"required" gorm:"column:usrupd"`
	// Flag                string    `json:"Flag"`
	DateUpd     time.Time `json:"dtmupd" gorm:"column:dtmupd"`
	IsSendEmail string    `json:"IsSendEmail"`
}
