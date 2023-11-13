package userController

import (
	"errors"
	"github.com/AhegaoHD/WBT/internal/models"
)

func validateStartTask(startTask *models.StartTaskRequest) error {
	if len(startTask.LoaderIDs) == 0 {
		return errors.New("len(startTask.LoaderIDs)==0")
	}
	return nil
}
