package model

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

// UserStatus represents the status of a user.
type UserStatus int32

const (
	UserStatusUnknown UserStatus = iota
	UserStatusNormal
	UserStatusDisabled
	UserStatusLocked

	DefaultUserStatus = UserStatusNormal
)

// IsValid checks if the status is valid
func (s UserStatus) IsValid() bool {
	return s == UserStatusNormal || s == UserStatusDisabled || s == UserStatusLocked
}
