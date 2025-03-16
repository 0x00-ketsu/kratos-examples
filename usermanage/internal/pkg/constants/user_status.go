package constants

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
