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

// paramSaveCase
// {
//     "TicketNo": "17283TEsFromGO",
//     "FlagCompany": "cnaf",
//     "BranchID": "419",
//     "AgreementNo": "419230221201",
//     "ApplicationID": "419A202309016922",
//     "CustomerID": "41900021011",
//     "CustomerName": "MARFIL SUNDALANGI",
//     "PhoneNo": "081243379372",
//     "Email": "baboljustitia@gmail.com",
//     "StatusID": "2",
//     "statusname" : "New",
//     "TypeID": "1",
//     "SubtypeID": "25",
//     "PriorityID": "1",
//     "Description": "fromGoFix",
//     "usrupd": "8023",
//     "ContactID": "1",
//     "RelationID": "1",
//     "RelationName": "Jane Doe",
//     "CallerID": "8080",
//     "Email_": "support@example.com",
//     "date_cr": "2025-01-01",
//     "foragingdays": "2025-01-01T00:00:00Z"
// }
