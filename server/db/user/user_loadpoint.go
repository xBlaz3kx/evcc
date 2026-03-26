package user

import (
	"time"

	"github.com/evcc-io/evcc/server/db"
	"gorm.io/gorm"
)

// UserLoadpoint links a user to loadpoint names they have access to.
// LoadpointName matches the config name used in loadpoint definitions.
type UserLoadpoint struct {
	ID            uint      `json:"id"            gorm:"primarykey;autoIncrement"`
	UserID        uint      `json:"userId"        gorm:"not null;index"`
	LoadpointName string    `json:"loadpointName" gorm:"not null;size:255"`
	CreatedAt     time.Time `json:"createdAt"`
}

func init() {
	db.Register(func(d *gorm.DB) error {
		return d.AutoMigrate(new(UserLoadpoint))
	})
}
