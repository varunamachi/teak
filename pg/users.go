package pg

import (
	"errors"

	"github.com/varunamachi/teak"
	"gopkg.in/hlandau/passlib.v1"
)

//userStorage - mongodb storage for user information
type userStorage struct{}

//NewUserStorage - creates a new user storage based on postgres
func NewUserStorage() teak.UserStorage {
	return &userStorage{}
}

//CreateUser - creates user in database
func (m *userStorage) CreateUser(user *teak.User) (err error) {
	query := `
		INSERT INTO teak_user(
			id,
			email,
			auth,
			firstName,
			lastName,
			title,
			fullName,
			state,
			verID,
			pwdExpiry,
			createdAt,
			createdBy,
			modifiedAt,
			modifiedBy,
			verified,
			props
		) VALUES (
			:id
			:email
			:auth
			:firstName
			:lastName
			:title
			:fullName
			:state
			:verID
			:pwdExpiry
			:createdAt
			:createdBy
			:modifiedAt
			:modifiedBy
			:verified
			:props
		)
	`
	_, err = db.NamedExec(query, user)
	return teak.LogError("t.user.pg", err)
}

//UpdateUser - updates user in database
func (m *userStorage) UpdateUser(user *teak.User) (err error) {
	query := `
		UPDATE teak_user SET 
			email = :email,
			auth = :auth,
			firstName = :firstName,
			lastName = :lastName,
			title = :title,
			fullName = :fullName,
			state = :state,
			verID = :verID,
			pwdExpiry = :pwdExpiry,
			createdAt = :createdAt,
			createdBy = :createdBy,
			modifiedAt = :modifiedAt,
			modifiedBy = :modifiedBy,
			verified = :verified,
			props = :props
		WHERE id = :id
	`
	_, err = db.NamedExec(query, user)
	return teak.LogError("t.user.pg", err)
}

//DeleteUser - deletes user with given user ID
func (m *userStorage) DeleteUser(userID string) (err error) {
	query := `DELETE FROM teak_user WHERE id = ?`
	_, err = db.Exec(query, userID)
	return teak.LogError("t.user.pg", err)
}

//GetUser - gets details of the user corresponding to ID
func (m *userStorage) GetUser(userID string) (user *teak.User, err error) {
	user = &teak.User{}
	query := `SELECT * FROM teak_user WHERE id = ?`
	db.Select(user, query, userID)
	return user, teak.LogError("t.user.pg", err)
}

//GetUsers - gets all users based on offset, limit and filter
func (m *userStorage) GetUsers(offset, limit int, filter *teak.Filter) (
	users []*teak.User, err error) {
	users = make([]*teak.User, 0, limit)
	selector := generateSelector(filter)
	query := `SELECT * FROM teak_users ` + selector
	err = db.Select(users, query, nil)
	return users, teak.LogError("t.user.pg", err)
}

//GetCount - gives the number of user selected by given filter
func (m *userStorage) GetCount(filter *teak.Filter) (count int, err error) {
	selector := generateSelector(filter)
	query := `SELECT COUNT(*) FROM teak_user ` + selector
	err = db.Select(&count, query)
	return count, teak.LogError("t.user.pg", err)
}

//GetUsersWithCount - Get users with total count
func (m *userStorage) GetUsersWithCount(
	offset, limit int, filter *teak.Filter) (
	total int, users []*teak.User, err error) {
	defer func() {
		teak.LogErrorX("t.user.pg",
			"Error getting count and list", err)
	}()
	selector := generateSelector(filter)
	get := `SELECT * FROM teak_user ` + selector
	count := `SELECT COUNT(*) FROM teak_user ` + selector
	users = make([]*teak.User, 0, limit)
	err = db.Select(users, get, nil)
	if err != nil {
		return
	}
	err = db.Select(&total, count, nil)
	return total, users, err
}

//ResetPassword - sets password of a unauthenticated user
func (m *userStorage) ResetPassword(
	userID, oldPwd, newPwd string) (err error) {
	defer func() {
		teak.LogError("t.user.pg", err)
	}()
	if err != nil {
		return err
	}
	// newHash, err := passlib.Hash(newPwd)
	_, err = passlib.Hash(newPwd)
	if err != nil {
		return err
	}
	if err = m.ValidateUser(userID, oldPwd); err != nil {
		err = errors.New("Could not match old password")
		return err
	}
	return teak.LogError("t.user.pg", err)
}

//SetPassword - sets password of a already authenticated user, old password
//is not required
func (m *userStorage) SetPassword(userID, newPwd string) (err error) {
	return teak.LogError("t.user.pg", err)
}

//ValidateUser - validates user ID and password
func (m *userStorage) ValidateUser(userID, password string) (err error) {
	return teak.LogError("t.user.pg", err)
}

//GetUserAuthLevel - gets user authorization level
func (m *userStorage) GetUserAuthLevel(
	userID string) (level teak.AuthLevel, err error) {
	return level, teak.LogError("t.user.pg", err)
}

//CreateSuperUser - creates the first super user for the application
func (m *userStorage) CreateSuperUser(
	user *teak.User, password string) (err error) {
	return teak.LogError("t.user.pg", err)
}

//SetUserState - sets state of an user account
func (m *userStorage) SetUserState(
	userID string, state teak.UserState) (err error) {
	return teak.LogError("t.user.pg", err)
}

//VerifyUser - sets state of an user account to verified based on userID
//and verification ID
func (m *userStorage) VerifyUser(userID, verID string) (err error) {
	return teak.LogError("t.user.pg", err)
}

//CreateIndices - creates mongoDB indeces for tables used for user management
func (m *userStorage) CreateIndices() (err error) {
	return err
}

//CleanData - cleans user management related data from database
func (m *userStorage) CleanData() (err error) {
	return err
}

//UpdateProfile - updates user details - this should be used when user logged in
//is updating own user account
func (m *userStorage) UpdateProfile(user *teak.User) (err error) {
	return teak.LogError("t.user.pg", err)
}
