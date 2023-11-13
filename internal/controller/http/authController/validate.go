package authController

import (
	"errors"
	"github.com/AhegaoHD/WBT/internal/models"
	"regexp"
)

func validateRegister(req *models.User) error {
	if req.UserType != "customer" && req.UserType != "loader" {
		return errors.New("req.UserType!=\"customer\" && req.UserType!=\"loader\"")
	}
	re := regexp.MustCompile("^[\\p{L}\\p{P}]+\\n*$")
	if !re.MatchString(req.Username) {
		return errors.New("user name contains invalid characters")
	}
	if !re.MatchString(req.Password) {
		return errors.New("password contains invalid characters")
	}
	return nil
}
