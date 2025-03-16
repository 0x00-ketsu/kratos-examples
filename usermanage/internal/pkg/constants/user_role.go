package constants

// UserRole represents the role of a user.
type UserRole int32

const (
	UserRoleUnknown UserRole = iota
	UserRoleAdmin
	UserRoleUser

	DefaultUserRole = UserRoleUser
)

// IsValid checks if the user role is valid.
func (r UserRole) IsValid() bool {
	return r == UserRoleUser || r == UserRoleAdmin
}
