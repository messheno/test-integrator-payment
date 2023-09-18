package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gobeam/stringy"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// APIGetAll
// @Summary      APIGetAll
// @Description  Recuperation de tous les utilisateurs
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/ [get]
func (u *UserModel) APIGetAll(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		users := []UserModel{}

		limit, _ := c.Get("LIMIT").(int64)
		if limit == 0 {
			limit = 10
		}
		offset, _ := c.Get("OFFSET").(int64)
		query, _ := c.Get("QUERY").(string)

		var count int64
		result := db.Model(&UserModel{}).Count(&count)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		result = db.Limit(int(limit)).Offset(int(offset)).Find(&users)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		fmt.Printf("USERS: %v", users)

		resp.SetData(map[string]interface{}{
			"users": users,
			"pagination": PaginationModel{
				Count:  count,
				Limit:  int(limit),
				Offset: int(offset),
				Query:  query,
			},
		})

		return resp.Send(c)
	}
}

// APICreate
// @Summary      APICreate
// @Description  Creation d'un nouvel utilisateur
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/ [post]
func (u *UserModel) APICreate(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(*JwtCustomClaims)
		if !ok {
			resp.SetStatus(http.StatusUnauthorized)
			err := fmt.Errorf("utilisateur non valide")
			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		u := new(UserModel)
		if err := c.Bind(u); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		if err := c.Validate(u); err != nil {
			log.Error().Err(err).Msgf("")

			errorsApi := ResErrorAPI{}

			// Traitement de la reponse de validator
			for _, err := range err.(validator.ValidationErrors) {
				str := stringy.New(err.Field())
				errorsApi = append(errorsApi, ErrorAPI{
					Code:    "400",
					Message: str.SnakeCase().ToLower(),
					Data:    err.Tag(),
				})
			}

			return resp.SendError(c, "Formulaire invalide", TransformErr(errorsApi))
		}

		if len(u.Password) <= 0 {
			err := fmt.Errorf("mot de passe obligatoire")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		// Vérification des autorisations du nouvelle utilisateur
		if u.isGrant(USER_ADMIN) && !claims.IsAdmin {
			err := fmt.Errorf("vous n'etes pas autorisé à crée un utilisateur admin")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		// Création de l'utilisateur
		result := db.Create(&u)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			if strings.Contains(result.Error.Error(), "duplicate key value violates") {
				return resp.SendError(c, "Existe déjà en base de donnée", TransformErr(result.Error))
			}

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		resp.SetData(map[string]interface{}{
			"user":        u,
			"user_create": true,
		})

		return resp.Send(c)
	}
}

// APIRead
// @Summary      APIRead
// @Description  Lecture des informations d'un utilisateur
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/:id [get]
func (u *UserModel) APIRead(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		user, ok := c.Get("USER").(*UserModel)
		if !ok {
			err := fmt.Errorf("utilisateur non valide")
			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		resp.SetData(map[string]interface{}{
			"user":      user,
			"user_read": true,
		})

		return resp.Send(c)
	}
}

// APIUpdate
// @Summary      APIUpdate
// @Description  Mise à jours des informations d'un utilisateur
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/:id [put]
func (u *UserModel) APIUpdate(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(*JwtCustomClaims)
		if !ok {
			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, "utilisateur non valide", TransformErr(fmt.Errorf("utilisateur non valide")))
		}

		// Récuperation jwt
		user, ok := c.Get("USER").(*UserModel)
		if !ok {
			return resp.SendError(c, "utilisateur non valide", TransformErr(fmt.Errorf("utilisateur non valide")))
		}

		// Récuperation des informations du formulaires
		u := new(UserModel)
		if err := c.Bind(u); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		u.CreatedAt = user.CreatedAt

		// Convert to json
		userFormJson, err := json.Marshal(u)
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		err = json.Unmarshal(userFormJson, user)
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		if err := c.Validate(user); err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, "Formulaire invalide", TransformErr(err))
		}

		// Vérification des autorisations du nouvelle utilisateur
		if user.isGrant(USER_ADMIN) && !claims.IsAdmin {
			err := fmt.Errorf("vous n'etes pas autorisé à modifié cette permission")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		// Création de l'utilisateur
		result := db.Save(user)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			if strings.Contains(result.Error.Error(), "duplicate key value violates") {
				return resp.SendError(c, "Existe déjà en base de donnée", TransformErr(result.Error))
			}

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		resp.SetData(map[string]interface{}{
			"user":        user,
			"user_update": true,
		})

		return resp.Send(c)
	}
}

// APIDelete
// @Summary      APIDelete
// @Description  Suppression d'un utilisateur
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/:id [delete]
func (u *UserModel) APIDelete(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		claims, ok := c.Get("JWT_CLAIMS").(*JwtCustomClaims)
		if !ok {
			resp.SetStatus(http.StatusUnauthorized)
			return resp.SendError(c, "utilisateur non valide", TransformErr(fmt.Errorf("utilisateur non valide")))
		}

		user, ok := c.Get("USER").(*UserModel)
		if !ok {
			return resp.SendError(c, "utilisateur non valide", TransformErr(fmt.Errorf("utilisateur non valide")))
		}

		// Vérification des autorisations du nouvelle utilisateur
		if user.isGrant(USER_ADMIN) || !claims.IsAdmin {
			err := fmt.Errorf("vous n'etes pas autorisé à supprimer cette utilisateur")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		result := db.Delete(user)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		resp.SetData(map[string]interface{}{
			"user":        user,
			"user_delete": true,
		})

		resp.Message = "Utilisateur supprimer avec succès"
		return resp.Send(c)
	}
}

// GetOnMid
func (u *UserModel) GetOnMid(db *gorm.DB) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
			if !ok {
				resp = NewResponseAPI[interface{}]()
			}

			// Récuperation jwt
			claims, ok := c.Get("JWT_CLAIMS").(*JwtCustomClaims)
			if !ok {
				resp.SetStatus(http.StatusUnauthorized)
				return resp.SendError(c, "utilisateur non valide", TransformErr(fmt.Errorf("utilisateur non valide")))
			}

			user := UserModel{}
			idValid := false

			// me
			if c.Param("id") == "me" {
				user.ID = claims.ID
				idValid = true
			} else if id, err := uuid.FromString(c.Param("id")); err == nil {
				user.ID = id.String()
				idValid = true
			}

			// email
			re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

			if !idValid && re.MatchString(c.Param("id")) {
				user.Email = c.Param("id")
				idValid = true
			}

			// number
			if !idValid && len(c.Param("id")) >= 8 {
				user.PhoneNumber = c.Param("id")
				idValid = true
			}

			if !idValid {
				return resp.SendError(c, "Identifiant utilisateur invalide doit id ou email ou numero", TransformErr(fmt.Errorf("identifiant utilisateur invalide doit id ou email ou numero")))
			}

			result := db.Where(&user).First(&user)
			if result.Error != nil {
				log.Error().Err(result.Error).Msgf("")

				return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
			}

			fmt.Println("SET USER")
			c.Set("USER", &user)

			if err := next(c); err != nil {
				// Retourne une erreur
				// c.Error(err)
				return resp.SendError(c, "Une erreur c'est produite", TransformErr(err))
			}

			return nil
		}
	}
}

// APILogin
// @Summary      APILogin
// @Description  Connexion
// @Tags         users
// @Accept       mpfd
// @Router       /api/users/login [post]
func (u *UserModel) APILogin(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, ok := c.Get("RESP").(*ResponseAPI[interface{}])
		if !ok {
			resp = NewResponseAPI[interface{}]()
		}

		// Form manager
		if len(c.FormValue("password")) <= 0 {
			err := fmt.Errorf("email/mobile ou mot de passe invalide")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		// Récuperation de l'utilsateur
		user := UserModel{}

		// email
		re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

		// phone
		if len(c.FormValue("phone_indicator")) > 0 && len(c.FormValue("phone_number")) > 0 && len(c.FormValue("phone_number")) >= 10 {
			user.PhoneIndicator = c.FormValue("phone_indicator")
			user.PhoneNumber = c.FormValue("phone_number")
		} else if len(c.FormValue("email")) > 0 && re.MatchString(c.FormValue("email")) {
			user.Email = c.FormValue("email")
		} else {
			err := fmt.Errorf("email/mobile ou mot de passe invalide")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		result := db.Where(&user).First(&user)
		if result.Error != nil {
			log.Error().Err(result.Error).Msgf("")

			if strings.Contains(result.Error.Error(), "record not found") {
				return resp.SendError(c, "Utilisateur ou mot de passe invalide", TransformErr(result.Error))
			}

			return resp.SendError(c, result.Error.Error(), TransformErr(result.Error))
		}

		// Check password
		ok, err := user.CheckPass(c.FormValue("password"))
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		if !ok {
			err := fmt.Errorf("username ou mot de passe invalide")
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		token, err := user.GenerateJWT()
		if err != nil {
			log.Error().Err(err).Msgf("")

			return resp.SendError(c, err.Error(), TransformErr(err))
		}

		resp.SetData(map[string]interface{}{
			"user_login": true,
			"token":      token,
			"token_type": "Baerer",
			"user":       user,
		})

		return resp.Send(c)
	}
}
