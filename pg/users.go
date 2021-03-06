package pg

import (
	"context"
	"time"

	"github.com/varunamachi/teak"
	"gopkg.in/hlandau/passlib.v1"
)

//userStorage - mongodb storage for user information
type userStorage struct{}

//NewUserStorage - creates a new user storage based on postgres
func NewUserStorage() teak.UserStorage {
	return &userStorage{}
	// TODO - make dataStorage satisfy teak.UserStorage interface
	// return nil
}

//CreateUser - creates user in database
func (m *userStorage) CreateUser(
	gtx context.Context, user *teak.User) (idHash string, err error) {
	if err = m.validateForSuper(gtx, user.Auth); err != nil {
		return "", err
	}
	if err = teak.UpdateUserInfo(user); err != nil {
		err = teak.LogErrorX("t.user.pg",
			"Failed to create user, user storage not properly configured", err)
		return "", err
	}
	query := `
		INSERT INTO teak_user(
			id,
			email,
			auth,
			first_name,
			last_name,
			title,
			full_name,
			state,
			ver_id,
			pwd_expiry,
			created_at,
			created_by,
			modified_at,
			modified_by,
			verified_at
		) VALUES (
			:id,
			:email,
			:auth,
			:first_name,
			:last_name,
			:title,
			:full_name,
			:state,
			:ver_id,
			:pwd_expiry,
			:created_at,
			:created_by,
			:modified_at,
			:modified_by,
			:verified_at
		)
	`
	//Skipped props for now
	_, err = defDB.NamedExecContext(gtx, query, user)
	return user.UserID, teak.LogError("t.user.pg", err)
}

//UpdateUser - updates user in database
func (m *userStorage) UpdateUser(
	gtx context.Context, user *teak.User) (err error) {
	if err = m.validateForSuper(gtx, user.Auth); err != nil {
		return err
	}
	query := `
		UPDATE teak_user SET 
			email = :email,
			auth = :auth,
			first_name = :first_name,
			last_name = :last_name,
			title = :title,
			full_name = :full_name,
			state = :state,
			ver_id = :ver_id,
			pwd_expiry = :pwd_expiry,
			created_at = :created_at,
			created_by = :created_by,
			modified_at = :modified_at,
			modified_by = :modified_by,
			verified_at = :verified_at,
			props = :props
		WHERE id = :id
	`
	_, err = defDB.NamedExecContext(gtx, query, user)
	return teak.LogError("t.user.pg", err)
}

//DeleteUser - deletes user with given user ID
func (m *userStorage) DeleteUser(
	gtx context.Context, userID string) (err error) {
	query := `DELETE FROM teak_user WHERE id = $1`
	_, err = defDB.ExecContext(gtx, query, userID)
	return teak.LogErrorX("t.user.pg", "Failed to delete user with id %s",
		err, userID)
}

//GetUser - gets details of the user corresponding to ID
func (m *userStorage) GetUser(
	gtx context.Context, userID string) (user *teak.User, err error) {
	user = &teak.User{}
	query := `SELECT * FROM teak_user WHERE id = $1`
	// defDB.Select(user, query, userID)
	err = defDB.GetContext(gtx, user, query, userID)
	return user, teak.LogError("t.user.pg", err)
}

//GetUsers - gets all users based on offset, limit and filter
func (m *userStorage) GetUsers(
	gtx context.Context, offset, limit int64, filter *teak.Filter) (
	users []*teak.User, err error) {
	users = make([]*teak.User, 0, limit)
	selector := generateSelector(filter)
	query := `SELECT * FROM teak_users ` + selector
	err = defDB.SelectContext(gtx, users, query, nil)
	return users, teak.LogError("t.user.pg", err)
}

//GetCount - gives the number of user selected by given filter
func (m *userStorage) GetCount(
	gtx context.Context, filter *teak.Filter) (count int64, err error) {
	selector := generateSelector(filter)
	query := `SELECT COUNT(*) FROM teak_user ` + selector
	err = defDB.SelectContext(gtx, &count, query)
	return count, teak.LogError("t.user.pg", err)
}

//GetUsersWithCount - Get users with total count
func (m *userStorage) GetUsersWithCount(
	gtx context.Context,
	offset, limit int64,
	filter *teak.Filter) (
	total int64, users []*teak.User, err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Error getting count and list", err)
	}()
	selector := generateSelector(filter)
	get := `SELECT * FROM teak_user ` + selector
	count := `SELECT COUNT(*) FROM teak_user ` + selector
	users = make([]*teak.User, 0, limit)
	err = defDB.SelectContext(gtx, users, get, nil)
	if err != nil {
		return
	}
	err = defDB.SelectContext(gtx, &total, count, nil)
	return total, users, err
}

//ResetPassword - sets password of a unauthenticated user
func (m *userStorage) ResetPassword(
	gtx context.Context,
	userID, oldPwd, newPwd string) (err error) {
	if err = m.ValidateUser(gtx, userID, oldPwd); err != nil {
		err = teak.LogErrorX("t.user.pg",
			"Reset password: Invalid current password given for userID %s",
			err, userID)
		return err
	}
	err = m.SetPassword(gtx, userID, newPwd)
	return err
}

//SetPassword - sets password of a already authenticated user, old password
//is not required
func (m *userStorage) SetPassword(
	gtx context.Context, userID, newPwd string) (err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Failed to set password for user %s", err, userID)
	}()
	newHash, err := passlib.Hash(newPwd)
	if err != nil {
		return err
	}
	//TODO - see why is new hash is set
	query := `
		INSERT INTO user_secret(user_id, phash) VALUES($1, $2)
			ON CONFLICT(user_id) DO UPDATE
				SET phash = EXCLUDED.phash
	`
	_, err = defDB.ExecContext(gtx, query, userID, newHash)
	return err
}

//ValidateUser - validates user ID and password
func (m *userStorage) ValidateUser(
	gtx context.Context, userID, password string) (err error) {
	defer func() {
		err = teak.LogErrorX("t.user.pg",
			"Failed to validate user with id %s", err, userID)
	}()
	var phash string
	err = defDB.GetContext(gtx, &phash,
		`SELECT phash FROM user_secret WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	newHash, err := passlib.Verify(password, phash)
	if err != nil {
		return err
	}
	if newHash != "" {
		query := `UPDATE user_secret SET phash = $1 WHERE user_id = $2`
		_, err = defDB.ExecContext(gtx, query, newHash, userID)
	}
	return err
}

//GetUserAuthLevel - gets user authorization level
func (m *userStorage) GetUserAuthLevel(
	gtx context.Context,
	userID string) (level teak.AuthLevel, err error) {
	err = defDB.GetContext(gtx, &level,
		`SELECT auth FROM teak_user WHERE id = $1`, userID)
	return level, teak.LogErrorX("t.user.pg",
		"Failed to retrieve auth level for '%s'", err, userID)
}

//SetAuthLevel - sets the auth level for the user
func (m *userStorage) SetAuthLevel(
	gtx context.Context,
	userID string,
	authLevel teak.AuthLevel) (err error) {
	if err = m.validateForSuper(gtx, authLevel); err != nil {
		return err
	}
	_, err = defDB.ExecContext(gtx,
		"UPDATE teak_user SET auth = $1 WHERE id = $2",
		userID, authLevel)
	return teak.LogErrorX("t.user.pg",
		"Failed to update auth level for user with ID '%s'", err, userID)
}

func (m *userStorage) validateForSuper(
	gtx context.Context,
	alevel teak.AuthLevel) (err error) {
	if alevel != teak.Super {
		return err //no error
	}
	numSuper := 0
	err = defDB.GetContext(gtx, &numSuper,
		"SELECT COUNT(*) FROM teak_user WHERE auth = 0")
	if err != nil {
		err = teak.LogErrorX("t.user.pg",
			"Failed to get number of super admins", err)
		return err
	}
	if numSuper >= 5 {
		err = teak.Error("t.user.pg",
			"Maximum limit for super admins reached")
		return err
	}
	return err
}

//SetUserState - sets state of an user account
func (m *userStorage) SetUserState(
	gtx context.Context,
	userID string,
	state teak.UserState) (err error) {
	_, err = defDB.ExecContext(gtx,
		"UPDATE teak_user SET state = $1 WHERE id = $2",
		userID, state)
	return teak.LogErrorX("t.user.pg",
		"Failed to update state for user with ID '%s'", err, userID)
}

//VerifyUser - sets state of an user account to verifiedAt based on userID
//and verification ID
func (m *userStorage) VerifyUser(
	gtx context.Context, userID, verID string) (err error) {
	query := `
		UPDATE teak_user SET 
			state = $1, 
			verified_at = $2, 
			ver_id = ""
		WHERE id = $3 AND ver_id = $4
	`
	_, err = defDB.ExecContext(
		gtx, query, teak.Active, time.Now(), userID, verID)
	return teak.LogErrorX("t.user.pg", "Failed to verify user with id %s",
		err, userID)
}

//CleanData - cleans user management related data from database
func (m *userStorage) CleanData(gtx context.Context) (err error) {
	_, err = defDB.ExecContext(gtx, `DELETE FROM teak_user`)
	return teak.LogErrorX("t.user.pg",
		"Failed to delete all user accounts", err)

}

//UpdateProfile - updates user details - this should be used when user logged in
//is updating own user account
func (m *userStorage) UpdateProfile(
	gtx context.Context, user *teak.User) (err error) {
	query := `
		UPDATE teak_user SET 
			email = $1,
			first_name = $2,
			last_name = $3,
			title = $4,
			full_name = $5,
			modified_at = $6,
			modified_by = $7
		WHERE id = $8
	`
	_, err = defDB.ExecContext(gtx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Title,
		user.FullName,
		time.Now(),
		user.FullName,
		user.UserID)
	return teak.LogError("t.user.pg", err)
}
