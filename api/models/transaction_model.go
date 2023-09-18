package models

type TransactionModel struct {
	Model

	ModeLive bool    `json:"mode_live" form:"mode_live" validate:"-"`
	Amount   float64 `json:"amount" form:"amount" validate:"required"`

	OperationMode  string `json:"operation_mode" form:"operation_mode" validate:"required"`   // CREDIT, DEBIT
	OperationState string `json:"operation_state" form:"operation_state" validate:"required"` // PENDING, SUCCESS, CANCEL, FAIL
	OperationMsg   string `json:"operation_msg" form:"operation_msg" validate:"-"`

	// Provider
	ProviderId string        `json:"provider_id" form:"provider_id" validate:"required" gorm:"index"`
	Provider   ProviderModel `json:"provider" form:"provider" validate:"required"`
}
