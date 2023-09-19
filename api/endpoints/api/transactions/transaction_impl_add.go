package transactions

import (
	"fmt"
	"spay/models"
	"spay/utils"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type ResTransactionAPICreateSuccess struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	IsError bool   `json:"is_error"`

	Data struct {
		Transaction models.TransactionModel `json:"transaction"`
	} `json:"data"`

	RequestDate time.Time `json:"request_date"`
	TimeElapsed string    `json:"time_elapsed"`
}

type AddFormData struct {
	Amount        float64 `json:"amount" form:"amount" xml:"amount" validate:"required"`                         // Montant de la transaction
	OperationMode string  `json:"operation_mode" form:"operation_mode" xml:"operation_mode" validate:"required"` // CREDIT, DEBIT
	ProviderId    string  `json:"provider_id" form:"provider_id" xml:"provider_id" validate:"required"`          // Identifiant du provider
	ReferenceId   string  `json:"reference_id" form:"reference_id" xml:"reference_id" validate:"required"`       // Reference de la transaction unique
	ServiceId     string  `json:"service_id" form:"service_id" xml:"service_id" validate:"required"`             // Reference de la transaction unique
	AuthId        string  `json:"auth_id" form:"auth_id" xml:"auth_id" validate:"omitempty"`                     // Executer en tant que
}

// Add
// @Summary      	Add new transaction
// @Description  	Création d'une nouvelle transaction
// @Tags         	Transactions
// @accept 			json,xml,x-www-form-urlencoded,mpfd
// @Product       	json
// @Param        	data formData AddFormData  false  "Contenu de la requete" ""
// @response      	200 {object} ResTransactionAPICreateSuccess
// @response      	400 {object} models.ResFailure
// @Router       	/api/transactions/ [post]
func (s *TransactionApiRessource) Add() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*models.ResponseAPI[interface{}])
		if !ok {
			resp = models.NewResponseAPI[interface{}]()
		}

		// Recuperation du claim de connexion
		claims, ok := c.Get("JWT_CLAIMS").(jwt.MapClaims)
		if !ok {
			err := fmt.Errorf("authentification obligatoire")
			// resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Récuperation des données du formulaire
		data := new(AddFormData)

		// Decodage de dataJson vers models.UserModel
		newTransaction := models.TransactionModel{}
		err := utils.BindValidate(c, data, &newTransaction)
		if err != nil {
			return resp.SendError(c, err.Error(), models.TransformErr(err))
		}

		// Création de la boutique
		_, err = createTransactionDb(&newTransaction, data.AuthId, claims)
		if err != nil {
			log.Error().Err(err).Msgf("")
			errMsg := err.Error()
			if strings.Contains(errMsg, "duplicated key not allowed") {
				errMsg = "La boutique existe déjà"
			}

			return resp.SendError(c, errMsg, models.TransformErr(err))
		}

		type resData struct {
			Transaction models.TransactionModel `json:"transaction"`
		}

		resp.SetData(resData{
			Transaction: newTransaction,
		})

		return resp.Send(c)
	}
}

func createTransactionDb(newTransaction *models.TransactionModel, authId string, claims jwt.MapClaims) (*models.UserModel, error) {
	// Connexion à la base de donnée
	db, err := models.GetDB()
	if err != nil {
		return nil, err
	}

	// Vérification des permissions
	loginUser := models.UserModel{}
	loginUser.AuthId = claims["sub"].(string)

	result := db.
		Preload("ServicePermissions").
		Where(&loginUser).First(&loginUser)
	if result.Error != nil {
		log.Error().Err(result.Error).Msgf("")
		if strings.Contains(result.Error.Error(), "record not found") {
			return nil, fmt.Errorf("utilisateur non reconnu")
		}

		return nil, result.Error
	}

	userAdmin := models.UserModel{}

	// // Si authId precisé
	if loginUser.IsGrant(models.USER_MANAGER) && authId != "" {
		result := db.Model(&userAdmin).Where("auth_id = ?", authId).First(&userAdmin)
		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		userAdmin = loginUser
	}

	// Provicer
	provider := models.ProviderModel{}
	db.Model(&provider).First(&provider)

	// Service

	result = db.Create(&newTransaction)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates") {
			return nil, fmt.Errorf("boutique existe déjà")
		}

		return nil, result.Error
	}

	return &userAdmin, nil
}
