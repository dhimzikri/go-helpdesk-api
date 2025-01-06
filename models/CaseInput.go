package models

import "time"

type Case struct {
	TicketNo      string    `json:"ticketno" binding:"required" gorm:"primaryKey;column:ticketno"`
	FlagCompany   string    `json:"flagcompany" binding:"required" gorm:"column:flagcompany"`
	BranchID      string    `json:"branchid" binding:"required" gorm:"column:branchid"`
	AgreementNo   string    `json:"agreementno" gorm:"column:agreementno"`
	ApplicationID string    `json:"applicationid" gorm:"column:applicationid"`
	CustomerID    string    `json:"customerid" gorm:"column:customerid"`
	CustomerName  string    `json:"customername" gorm:"column:customername"`
	PhoneNo       string    `json:"phoneno" gorm:"column:phoneno"`
	Email         string    `json:"email" gorm:"column:email"`
	StatusID      int       `json:"statusid" binding:"required" gorm:"column:statusid"`
	TypeID        int       `json:"typeid" gorm:"column:typeid"`
	SubtypeID     int       `json:"subtypeid" gorm:"column:subtypeid"`
	PriorityID    int       `json:"priorityid" gorm:"column:priorityid"`
	Description   string    `json:"description" gorm:"column:description"`
	UserID        string    `json:"usrupd" binding:"required" gorm:"column:usrupd"`
	ContactID     int       `json:"contactid" gorm:"column:contactid"`
	RelationID    int       `json:"relationid" gorm:"column:relationid"`
	RelationName  string    `json:"relationname" gorm:"column:relationname"`
	CallerID      int       `json:"callerid" gorm:"column:callerid"`
	Email_        string    `json:"email_" binding:"required" gorm:"column:email_"`
	DateCr        string    `json:"date_cr" binding:"required" gorm:"column:date_cr"`
	ForAgingDays  time.Time `json:"foragingdays" gorm:"column:foragingdays"`
	StatusDesc    string    `json:"statusname" gorm:"column:statusname"`
}
