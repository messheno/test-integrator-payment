package models

type TransactionModel struct {
	Model

	ModeLive      bool    `json:"mode_live" form:"mode_live" validate:"-"`
	Amount        float64 `json:"amount" form:"amount" validate:"required"`
	AmountWithFee float64 `json:"amount_with_fee" form:"amount_with_fee" validate:"required"`

	OperationMode  string `json:"operation_mode" form:"operation_mode" validate:"required"`   // CREDIT, DEBIT
	OperationState string `json:"operation_state" form:"operation_state" validate:"required"` // PENDING, SUCCESS, CANCEL, FAIL
	OperationMsg   string `json:"operation_msg" form:"operation_msg" validate:"-"`

	// Provider
	ProviderId string        `json:"provider_id" form:"provider_id" validate:"required" gorm:"index"`
	Provider   ProviderModel `json:"provider" form:"provider" validate:"required"`

	// Shop
	ShopId string    `json:"shop_id" form:"shop_id" validate:"required" gorm:"index"`
	Shop   ShopModel `json:"shop" form:"shop" validate:"required"`

	// Réference
	ReferenceId string `json:"reference_id" form:"reference_id" validate:"required"` // Id facture
	Currency    string `json:"currency" form:"currency" validate:"required"`
}
