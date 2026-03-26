package user

import (
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/server/db/settings"
	"gorm.io/gorm"
)

// MigrateAdminPassword migrates the legacy single-admin password to the users table.
// If the users table is empty and an adminPassword exists in settings, it creates an
// "admin" user with role admin, reusing the existing bcrypt hash directly.
// This runs as part of db.Register after AutoMigrate so existing installs are not broken.
func MigrateAdminPassword(d *gorm.DB) error {
	var count int64
	if err := d.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // users already exist, skip migration
	}

	hash, err := settings.String(keys.AdminPassword)
	if err != nil || hash == "" {
		return nil // no legacy password, fresh install
	}

	admin := User{
		Username:     "admin",
		PasswordHash: hash,
		Role:         RoleAdmin,
	}
	return d.Create(&admin).Error
}
