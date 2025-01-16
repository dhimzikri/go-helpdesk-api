package models

type CaseRequest struct {
	Task          string `json:"task" binding:"required"`
	IsSendEmail   string `json:"IsSendEmail"`
	AgreementNo   string `json:"AgreementNo"`
	ApplicationID string `json:"ApplicationID"`
	BranchID      string `json:"BranchID"`
	CallerID      string `json:"CallerID"`
	ContactID     int    `json:"ContactID"`
	CustomerID    string `json:"CustomerID"`
	CustomerName  string `json:"CustomerName"`
	DateCr        string `json:"DateCr"`
	Description   string `json:"Description"`
	Email         string `json:"Email"`
	Email_        string `json:"Email_"`
	FlagCompany   string `json:"FlagCompany"`
	PhoneNo       string `json:"PhoneNo"`
	PriorityID    int    `json:"PriorityID"`
	RelationID    int    `json:"RelationID"`
	RelationName  string `json:"RelationName"`
	StatusID      int    `json:"StatusID"`
	SubTypeID     int    `json:"SubTypeID"`
	TicketNo      string `json:"TicketNo"`
	TypeID        int    `json:"TypeID"`
	UserID        string `json:"UserID"`
	Flag          string `json:"Flag"`
}
