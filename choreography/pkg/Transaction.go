package pkg

type Transaction struct {
	TransactionID  int    `json:"transaction_id"`
	ServiceAStatus bool   `json:"service_a_status"`
	ServiceBStatus bool   `json:"service_b_status"`
	Sender         string `json:"sender"`
}
