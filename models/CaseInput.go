package models

import "time"

type Case struct {
	IsSendEmail         string    `json:"IsSendEmail"`
	AgreementNo         string    `json:"AgreementNo"`
	ApplicationID       string    `json:"ApplicationID"`
	BranchID            string    `json:"BranchID"`
	CallerID            string    `json:"CallerID"`
	ContactDescription  string    `json:"ContactDescription"`
	ContactID           int       `json:"ContactID"`
	CustomerID          string    `json:"CustomerID"`
	CustomerName        string    `json:"CustomerName"`
	DateCr              string    `json:"DateCr"`
	Description         string    `json:"Description"`
	Email               string    `json:"Email"`
	Email_              string    `json:"Email_"`
	FlagCompany         string    `json:"FlagCompany"`
	PhoneNo             string    `json:"PhoneNo"`
	PriorityDescription string    `json:"PriorityDescription"`
	PriorityID          int       `json:"PriorityID"`
	RelationDescription string    `json:"RelationDescription"`
	RelationID          int       `json:"RelationID"`
	RelationName        string    `json:"RelationName"`
	StatusDescription   *string   `json:"StatusDescription"`
	StatusID            int       `json:"StatusID"`
	StatusName          string    `json:"StatusName"`
	SubTypeID           int       `json:"SubTypeID"`
	TicketNo            string    `json:"TicketNo"`
	TypeDescription     string    `json:"TypeDescription"`
	TypeID              int       `json:"TypeID"`
	TypeSubDescription  string    `json:"TypeSubDescription"`
	UserID              string    `json:"UserID"`
	Flag                string    `json:"Flag"`
	DateUpd             time.Time `json:"dtmupd"`
}
