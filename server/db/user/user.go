package user

import (
	"time"

	"github.com/evcc-io/evcc/server/db"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role represents the user's RBAC role
type Role string

const (
	RoleViewer     Role = "viewer"
	RoleUser       Role = "user"
	RoleMaintainer Role = "maintainer"
	RoleAdmin      Role = "admin"
)

// User represents a system user
type User struct {
	ID           uint      `json:"id"        gorm:"primarykey;autoIncrement"`
	Username     string    `json:"username"  gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string    `json:"-"         gorm:"not null"`
	Role         Role      `json:"role"      gorm:"not null;default:'viewer'"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Vehicles     []string  `json:"vehicles"   gorm:"-"`
	Loadpoints   []string  `json:"loadpoints" gorm:"-"`
}

func init() {
	db.Register(func(d *gorm.DB) error {
		if err := d.AutoMigrate(new(User)); err != nil {
			return err
		}
		return MigrateAdminPassword(d)
	})
}

// loadVehicles populates the Vehicles slice for a single user
func loadVehicles(u *User) {
	var uvs []UserVehicle
	if err := db.Instance.Where("user_id = ?", u.ID).Find(&uvs).Error; err != nil {
		return
	}
	names := make([]string, 0, len(uvs))
	for _, uv := range uvs {
		names = append(names, uv.VehicleName)
	}
	u.Vehicles = names
}

// loadLoadpoints populates the Loadpoints slice for a single user
func loadLoadpoints(u *User) {
	var uls []UserLoadpoint
	if err := db.Instance.Where("user_id = ?", u.ID).Find(&uls).Error; err != nil {
		return
	}
	names := make([]string, 0, len(uls))
	for _, ul := range uls {
		names = append(names, ul.LoadpointName)
	}
	u.Loadpoints = names
}

// LoadpointsForUser returns loadpoint names attached to a user
func LoadpointsForUser(userID uint) ([]string, error) {
	var uls []UserLoadpoint
	if err := db.Instance.Where("user_id = ?", userID).Find(&uls).Error; err != nil {
		return nil, err
	}
	names := make([]string, 0, len(uls))
	for _, ul := range uls {
		names = append(names, ul.LoadpointName)
	}
	return names, nil
}

// AddLoadpoint attaches a loadpoint name to a user (idempotent)
func AddLoadpoint(userID uint, loadpointName string) error {
	ul := UserLoadpoint{UserID: userID, LoadpointName: loadpointName}
	return db.Instance.
		Where("user_id = ? AND loadpoint_name = ?", userID, loadpointName).
		FirstOrCreate(&ul).Error
}

// RemoveLoadpoint detaches a loadpoint name from a user
func RemoveLoadpoint(userID uint, loadpointName string) error {
	return db.Instance.
		Where("user_id = ? AND loadpoint_name = ?", userID, loadpointName).
		Delete(&UserLoadpoint{}).Error
}

// RemoveVehicleFromAllUsers removes a vehicle name from every user that had it assigned
func RemoveVehicleFromAllUsers(vehicleName string) error {
	return db.Instance.
		Where("vehicle_name = ?", vehicleName).
		Delete(&UserVehicle{}).Error
}

// RemoveLoadpointFromAllUsers removes a loadpoint name from every user that had it assigned
func RemoveLoadpointFromAllUsers(loadpointName string) error {
	return db.Instance.
		Where("loadpoint_name = ?", loadpointName).
		Delete(&UserLoadpoint{}).Error
}

// VehiclesForUser returns vehicle names attached to a user
func VehiclesForUser(userID uint) ([]string, error) {
	var uvs []UserVehicle
	if err := db.Instance.Where("user_id = ?", userID).Find(&uvs).Error; err != nil {
		return nil, err
	}
	names := make([]string, 0, len(uvs))
	for _, uv := range uvs {
		names = append(names, uv.VehicleName)
	}
	return names, nil
}

// AddVehicle attaches a vehicle name to a user (idempotent)
func AddVehicle(userID uint, vehicleName string) error {
	uv := UserVehicle{UserID: userID, VehicleName: vehicleName}
	return db.Instance.
		Where("user_id = ? AND vehicle_name = ?", userID, vehicleName).
		FirstOrCreate(&uv).Error
}

// RemoveVehicle detaches a vehicle name from a user
func RemoveVehicle(userID uint, vehicleName string) error {
	return db.Instance.
		Where("user_id = ? AND vehicle_name = ?", userID, vehicleName).
		Delete(&UserVehicle{}).Error
}

// Count returns the number of users in the database
func Count() (int64, error) {
	var count int64
	return count, db.Instance.Model(&User{}).Count(&count).Error
}

// AdminCount returns the number of admin users excluding the given user ID
func AdminCount(excludeID uint) (int64, error) {
	var count int64
	return count, db.Instance.Model(&User{}).Where("role = ? AND id != ?", RoleAdmin, excludeID).Count(&count).Error
}

// All returns all users with their attached vehicles and loadpoints
func All() ([]User, error) {
	var users []User
	if err := db.Instance.Find(&users).Error; err != nil {
		return nil, err
	}
	for i := range users {
		loadVehicles(&users[i])
		loadLoadpoints(&users[i])
	}
	return users, nil
}

// ByID returns a user by primary key with attached vehicles and loadpoints
func ByID(id any) (*User, error) {
	var u User
	if err := db.Instance.First(&u, id).Error; err != nil {
		return nil, err
	}
	loadVehicles(&u)
	loadLoadpoints(&u)
	return &u, nil
}

// ByUsername returns a user by username
func ByUsername(username string) (*User, error) {
	var u User
	return &u, db.Instance.Where("username = ?", username).First(&u).Error
}

// Create inserts a new user
func Create(u *User) error {
	return db.Instance.Create(u).Error
}

// Save persists changes to an existing user
func Save(u *User) error {
	return db.Instance.Save(u).Error
}

// Delete removes a user from the database
func Delete(u *User) error {
	return db.Instance.Delete(u).Error
}

// SetPassword hashes and stores the password using bcrypt
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies a plaintext password against the stored hash
func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
