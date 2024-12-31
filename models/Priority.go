package models

type Priority struct {
	PriorityID  int    `json:"priorityid"`
	Description string `json:"description"`
	Seq         int    `json:"seq"`
	UsrUpd      string `json:"usrupd"`
	DtmUpd      string `json:"dtmupd"`
}

type Response struct {
	Total   int        `json:"total"`
	Success bool       `json:"success"`
	Data    []Priority `json:"data"`
}
