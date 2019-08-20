package pg

import (
	"time"

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
			verifiedAt,
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
			:verifiedAt
			:props
		)
	`
	_, err = defDB.NamedExec(query, user)
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
			verifiedAt = :verifiedAt,
			props = :props
		WHERE id = :id
	`
	_, err = defDB.NamedExec(query, user)
	return teak.LogError("t.user.pg", err)
}

//DeleteUser - deletes user with given user ID
func (m *userStorage) DeleteUser(userID string) (err error) {
	query := `DELETE FROM teak_user WHERE id = ?`
	_, err = defDB.Exec(query, userID)
	return teak.LogErrorX("t.user.pg", "Failed to delete user with id %s",
		err, userID)
}

//GetUser - gets details of the user corresponding to ID
func (m *userStorage) GetUser(userID string) (user *teak.User, err error) {
	user = &teak.User{}
	query := `SELECT * FROM teak_user WHERE id = ?`
	defDB.Select(user, query, userID)
	return user, teak.LogError("t.user.pg", err)
}

//GetUsers - gets all users based on offset, limit and filter
func (m *userStorage) GetUsers(offset, limit int, filter *teak.Filter) (
	users []*teak.User, err error) {
	users = make([]*teak.User, 0, limit)
	selector := generateSelector(filter)
	query := `SELECT * FROM teak_users ` + selector
	err = defDB.Select(users, query, nil)
	return users, teak.LogError("t.user.pg", err)
}

//GetCount - gives the number of user selected by given filter
func (m *userStorage) GetCount(filter *teak.Filter) (count int, err error) {
	selector := generateSelector(filter)
	query := `SELECT COUNT(*) FROM teak_user ` + selector
	err = defDB.Select(&count, query)
	return count, teak.LogError("t.user.pg", err)
}

//GetUsersWithCount - Get users with total count
func (m *userStorage) GetUsersWithCount(
	offset, limit int, filter *teak.Filter) (
	total int, users []*teak.User, err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Error getting count and list", err)
	}()
	selector := generateSelector(filter)
	get := `SELECT * FROM teak_user ` + selector
	count := `SELECT COUNT(*) FROM teak_user ` + selector
	users = make([]*teak.User, 0, limit)
	err = defDB.Select(users, get, nil)
	if err != nil {
		return
	}
	err = defDB.Select(&total, count, nil)
	return total, users, err
}

//ResetPassword - sets password of a unauthenticated user
func (m *userStorage) ResetPassword(
	userID, oldPwd, newPwd string) (err error) {
	if err = m.ValidateUser(userID, oldPwd); err != nil {
		err = teak.LogErrorX("t.user.pg",
			"Reset password: Invalid current password given for userID %s",
			err, userID)
		return err
	}
	err = m.SetPassword(userID, newPwd)
	return err
}

//SetPassword - sets password of a already authenticated user, old password
//is not required
func (m *userStorage) SetPassword(userID, newPwd string) (err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Failed to set password for user %s", err, userID)
	}()
	newHash, err := passlib.Hash(newPwd)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO user_secret(userID, phash) VALUES(?, ?)
			ON CONFLICT DO UPDATE
				SET phash = EXCLUDED.phash
	`
	_, err = defDB.Exec(query, userID, newHash)
	return err
}

//ValidateUser - validates user ID and password
func (m *userStorage) ValidateUser(userID, password string) (err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Failed to validate user with id %s", err, userID)
	}()
	var phash string
	err = defDB.Select(&phash,
		`SELECT phash FROM user_secret WHERE userID = ?`, userID)
	if err != nil {
		return err
	}
	newHash, err := passlib.Verify(password, phash)
	if err != nil {
		return err
	}
	query := `UPDATE user_secret SET phash = ? WHERE userID = ?`
	_, err = defDB.Exec(query, newHash, userID)
	return err
}

//GetUserAuthLevel - gets user authorization level
func (m *userStorage) GetUserAuthLevel(
	userID string) (level teak.AuthLevel, err error) {
	err = defDB.Select(&level, `SELECT auth FROM teak_user WHERE id = ?`, userID)
	return level, teak.LogErrorX("t.user.pg",
		"Failed to retrieve auth level for '%s'", err, userID)
}

//CreateSuperUser - creates the first super user for the application
func (m *userStorage) CreateSuperUser(
	user *teak.User, password string) (err error) {
	///Todo change this interface to set role...
	///If role is super add extra validations

	// numSuper := 0
	// err = defDB.Select(&numSuper,
	// 		"SELECT COUNT(*) FROM teak_user WHERE auth = 0")
	// if err != nil {
	// 	err = teak.LogErrorX("t.user.pg",
	// 		"Failed to get number of super admins", err)
	// 	return err
	// }
	// if numSuper >= 5 {
	// 	err = teak.Error("t.user.pg", "Maximum limit for super admins reached")
	// 	return err
	// }
	// query := `UPDATE teak_user SET auth = 0 WHERE id = ?`
	// _, err = defDB.Exec(query, user.ID)
	// if err != nil {
	// 	err = teak.LogErrorX("t.user.pg",
	// 		"Failed to set super user role to %s", err, user.FullName)
	// }
	return err
}

//SetUserState - sets state of an user account
func (m *userStorage) SetUserState(
	userID string, state teak.UserState) (err error) {
	_, err = defDB.Exec("UPDATE teak_user SET state = ? WHERE id = ?",
		userID, state)
	return teak.LogErrorX("t.user.pg",
		"Failed to update state for user with ID '%s'", err, userID)
}

//VerifyUser - sets state of an user account to verifiedAt based on userID
//and verification ID
func (m *userStorage) VerifyUser(userID, verID string) (err error) {
	query := `
		UPDATE teak_user SET 
			state = ?, 
			verifiedAt = ?, 
			verID = ""
		WHERE id = ? AND verID = ?
	`
	_, err = defDB.Exec(query, teak.Active, time.Now(), userID, verID)
	return teak.LogErrorX("t.user.pg", "Failed to verify user with id %s",
		err, userID)
}

//CleanData - cleans user management related data from database
func (m *userStorage) CleanData() (err error) {
	_, err = defDB.Exec(`DELETE FROM teak_user`)
	return teak.LogErrorX("t.user.pg",
		"Failed to delete all user accounts", err)

}

//UpdateProfile - updates user details - this should be used when user logged in
//is updating own user account
func (m *userStorage) UpdateProfile(user *teak.User) (err error) {
	query := `
		UPDATE teak_user SET 
			email = ?,
			firstName = ?,
			lastName = ?,
			title = ?,
			fullName = ?,
			modifiedAt = ?,
			modifiedBy = ?
		WHERE id = ?
	`
	_, err = defDB.Exec(query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Title,
		user.FullName,
		time.Now(),
		user.FullName)
	return teak.LogError("t.user.pg", err)
}
