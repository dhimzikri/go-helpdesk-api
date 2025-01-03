package models

import "time"

type CaseInput struct {
	TicketNo      string    `json:"ticketno" binding:"required"`
	FlagCompany   string    `json:"flagcompany" binding:"required"`
	BranchID      string    `json:"branchid" binding:"required"`
	AgreementNo   string    `json:"agreementno"`
	ApplicationID string    `json:"applicationid"`
	CustomerID    string    `json:"customerid"`
	CustomerName  string    `json:"customername"`
	PhoneNo       string    `json:"phoneno"`
	Email         string    `json:"email"`
	StatusID      string    `json:"statusid" binding:"required"`
	TypeID        string    `json:"typeid"`
	SubtypeID     string    `json:"subtypeid"`
	PriorityID    string    `json:"priorityid"`
	Description   string    `json:"description"`
	UserID        string    `json:"usrupd" binding:"required"`
	ContactID     string    `json:"contactid"`
	RelationID    string    `json:"relationid"`
	RelationName  string    `json:"relationname"`
	CallerID      string    `json:"callerid"`
	Email_        string    `json:"email_"`
	DateCr        string    `json:"date_cr"`
	ForAgingDays  time.Time `json:"foragingdays"`
	StatusDesc    string    `json:"statusname"`
}
