// models/case_model.go
package models

// Case represents a single case record
type Case struct {
	TicketNo     string `json:"ticketno"`
	AgreementNo  string `json:"agreementno"`
	CustomerName string `json:"customername"`
	StatusName   string `json:"statusname"`
	StatusDesc   string `json:"statusdescription"`
	// Add other fields as needed
}

// CaseResponse represents the response structure for case data
type CaseResponse struct {
	Total   int    `json:"total"`
	Success bool   `json:"success"`
	Cases   []Case `json:"data"`
}
