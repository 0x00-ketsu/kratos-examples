package biz

// TODO: read it from config or env.
// DefaultUserPassword is the default password for a user.
const DefaultUserPassword = "Jv#r7SXUlQ8r^EdQ"

// UserRole represents the role of a user.
type UserRole int32

const (
	UserRoleUnknown UserRole = iota
	UserRoleAdmin
	UserRoleUser
)

// UserStatus represents the status of a user.
type UserStatus int32

const (
	UserStatusUnknown UserStatus = iota
	UserStatusNormal
	UserStatusDisabled
	UserStatusLocked
)

const (
	DefaultUserRole   = UserRoleUser
	DefaultUserStatus = UserStatusNormal
)

// String returns the string repetition of the user role.
func (r UserRole) String() string {
	switch r {
	case UserRoleAdmin:
		return "admin"
	case UserRoleUser:
		return "user"
	default:
		return "unknown"
	}
}

// IsValid checks if the user role is valid.
func (r UserRole) IsValid() bool {
	return r == UserRoleAdmin || r == UserRoleUser
}

// String returns the string repetition of the user status.
func (s UserStatus) String() string {
	switch s {
	case UserStatusNormal:
		return "normal"
	case UserStatusDisabled:
		return "disabled"
	case UserStatusLocked:
		return "locked"
	default:
		return "unknown"
	}
}

// IsValid checks if the user status is valid.
func (s UserStatus) IsValid() bool {
	return s == UserStatusNormal || s == UserStatusDisabled || s == UserStatusLocked
}

// IsNormal checks if the user status is normal.
func (s UserStatus) IsNormal() bool {
	return s == UserStatusNormal
}
