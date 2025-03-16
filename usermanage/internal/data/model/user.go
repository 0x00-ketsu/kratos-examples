package model

import (
	"usermanage/internal/pkg/constants"
	"usermanage/internal/pkg/id"
	"usermanage/internal/pkg/password"

	"gorm.io/gorm"
)

// User represents the user entity.
type User struct {
	BaseModel
	Username string             `json:"username" gorm:"index;size:64"`
	Password string             `json:"password" gorm:"size:128"`
	Role     constants.UserRole `json:"role"`
	Status   constants.UserStatus         `json:"status"`
	// MustChangePassword indicates if the user must change the password.
	// The default value is `true`.
	// MustChangePassword bool   `json:"mustChangePassword" gorm:"default:true"`
	Creator   string `json:"creator" gorm:"size:64"`
	UpdatedBy string `json:"updatedBy" gorm:"size:64"`
}

// BeforeCreate a Gorm hook to be run before the user is created.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = id.GenerateUUID(true)
	u.SetPassword(u.Password)
	if !u.Role.IsValid() {
		u.Role = constants.DefaultUserRole
	}
	if !u.Status.IsValid() {
		u.Status = constants.DefaultUserStatus
	}
	return
}

// VerifyPassword compares a hashed password with a raw password.
func (u *User) VerifyPassword(rawPassword string) bool {
	return password.Verify(u.Password, rawPassword)
}

// SetPassword sets the password for the user.
func (u *User) SetPassword(rawPassword string) error {
	hash, err := password.Hash(rawPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}
