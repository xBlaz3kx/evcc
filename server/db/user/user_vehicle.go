package user

import (
	"time"

	"github.com/evcc-io/evcc/server/db"
	"gorm.io/gorm"
)

// UserVehicle links a user to vehicle names they own.
// VehicleName matches the config name used in loadpoint vehicle assignments.
type UserVehicle struct {
	ID          uint      `json:"id"          gorm:"primarykey;autoIncrement"`
	UserID      uint      `json:"userId"      gorm:"not null;index"`
	VehicleName string    `json:"vehicleName" gorm:"not null;size:255"`
	CreatedAt   time.Time `json:"createdAt"`
}

func init() {
	db.Register(func(d *gorm.DB) error {
		return d.AutoMigrate(new(UserVehicle))
	})
}
