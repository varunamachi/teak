package mongo

import (
	"errors"
	"time"

	"github.com/varunamachi/teak"
	"gopkg.in/hlandau/passlib.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//userStorage - mongodb storage for user information
type userStorage struct{}

//NewUserStorage - creates a new user storage based on mongodb
func NewUserStorage() teak.UserStorage {
	return &userStorage{}
}

//CreateUser - creates user in database
func (m *userStorage) CreateUser(user *teak.User) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	if err = m.validateForSuper(conn, user.Auth); err != nil {
		return err
	}
	err = conn.C("users").Insert(user)
	return teak.LogError("t.user.mongo", err)
}

//UpdateUser - updates user in database
func (m *userStorage) UpdateUser(user *teak.User) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	if err = m.validateForSuper(conn, user.Auth); err != nil {
		return err
	}
	err = conn.C("users").Update(bson.M{"id": user.ID}, user)
	return teak.LogError("t.user.mongo", err)
}

//DeleteUser - deletes user with given user ID
func (m *userStorage) DeleteUser(userID string) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C("users").Remove(bson.M{"id": userID})
	return teak.LogError("t.user.mongo", err)
}

//GetUser - gets details of the user corresponding to ID
func (m *userStorage) GetUser(userID string) (user *teak.User, err error) {
	conn := DefaultConn()
	user = &teak.User{}
	defer conn.Close()
	err = conn.C("users").Find(bson.M{"id": userID}).One(user)
	return user, teak.LogError("t.user.mongo", err)
}

//GetUsers - gets all users based on offset, limit and filter
func (m *userStorage) GetUsers(offset, limit int, filter *teak.Filter) (
	users []*teak.User, err error) {
	conn := DefaultConn()
	defer conn.Close()
	selector := generateSelector(filter)
	users = make([]*teak.User, 0, limit)
	err = conn.C("users").
		Find(selector).
		Sort("-created").
		Skip(offset).
		Limit(limit).
		All(&users)
	return users, teak.LogError("t.user.mongo", err)
}

//GetCount - gives the number of user selected by given filter
func (m *userStorage) GetCount(filter *teak.Filter) (count int, err error) {
	conn := DefaultConn()
	defer conn.Close()
	selector := generateSelector(filter)
	count, err = conn.C("users").Find(selector).Count()
	return count, teak.LogError("t.user.mongo", err)
}

//GetUsersWithCount - Get users with total count
func (m *userStorage) GetUsersWithCount(
	offset, limit int, filter *teak.Filter) (
	total int, users []*teak.User, err error) {
	conn := DefaultConn()
	defer conn.Close()
	var selector bson.M
	selector = generateSelector(filter)
	users = make([]*teak.User, 0, limit)
	q := conn.C("users").Find(selector).Sort("-created")
	total, err = q.Count()
	if err == nil {
		err = q.Skip(offset).Limit(limit).All(&users)
	}
	return total, users, teak.LogError("t.user.mongo", err)
}

//ResetPassword - sets password of a unauthenticated user
func (m *userStorage) ResetPassword(
	userID, oldPwd, newPwd string) (err error) {
	conn := DefaultConn()
	defer func() {
		conn.Close()
		teak.LogError("t.user.mongo", err)
	}()
	if err != nil {
		return err
	}
	var newHash string
	newHash, err = passlib.Hash(newPwd)
	if err != nil {
		return err
	}
	if err = m.ValidateUser(userID, oldPwd); err != nil {
		err = errors.New("Could not match old password")
		return err
	}
	err = conn.C("secret").Update(
		bson.M{
			"userID": userID,
		},
		bson.M{
			"$set": bson.M{
				"phash": newHash,
			},
		},
	)
	return teak.LogError("UMan:Mongo", err)
}

//SetPassword - sets password of a already authenticated user, old password
//is not required
func (m *userStorage) SetPassword(userID, newPwd string) (err error) {
	var newHash string
	newHash, err = passlib.Hash(newPwd)
	if err == nil {
		conn := DefaultConn()
		defer conn.Close()
		m.setPasswordHash(conn, userID, newHash)
	}
	return teak.LogError("UMan:Mongo", err)
}

func (m *userStorage) setPasswordHash(conn *Conn,
	userID, hash string) (err error) {
	_, err = conn.C("secret").Upsert(
		bson.M{
			"userID": userID,
		},
		bson.M{
			"$set": bson.M{
				"userID": userID,
				"phash":  hash,
			},
		})
	return err
}

//ValidateUser - validates user ID and password
func (m *userStorage) ValidateUser(userID, password string) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	secret := bson.M{}
	err = conn.C("secret").
		Find(bson.M{"userID": userID}).
		Select(bson.M{"phash": 1, "_id": 0}).
		One(&secret)
	if err == nil {
		storedHash, ok := secret["phash"].(string)
		if ok {
			var newHash string
			newHash, err = passlib.Verify(password, storedHash)
			if err == nil && newHash != "" {
				err = m.setPasswordHash(conn, userID, newHash)
			}
		} else {
			err = errors.New("Failed to varify password")
		}
	}
	return teak.LogError("UMan:Mongo", err)
}

//GetUserAuthLevel - gets user authorization level
func (m *userStorage) GetUserAuthLevel(
	userID string) (level teak.AuthLevel, err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C("users").
		Find(bson.M{"userID": userID}).
		Select(bson.M{"auth": 1}).
		One(&level)
	return level, teak.LogError("UMan:Mongo", err)
}

//SetAuthLevel - sets the auth level for the user
func (m *userStorage) SetAuthLevel(
	userID string, authLevel teak.AuthLevel) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	if err = m.validateForSuper(conn, authLevel); err != nil {
		return err
	}
	err = conn.C("users").Update(
		bson.M{
			"id": userID,
		},
		bson.M{
			"$set": bson.M{
				"auth": authLevel,
			},
		},
	)
	return teak.LogError("t.user.mongo", err)
}

func (m *userStorage) validateForSuper(
	conn *Conn, alevel teak.AuthLevel) (err error) {
	if alevel != teak.Super {
		return err //no error
	}
	var count int
	count, err = conn.C("users").Find(bson.M{"auth": 0}).Count()
	if err != nil {
		err = teak.LogErrorX("t.user.mongo",
			"Failed to get number of super admins", err)
		return err
	}
	if count > 5 {
		err = teak.Error("t.user.mongo",
			"Maximum limit for super admins reached")
		return err
	}
	return err
}

//SetUserState - sets state of an user account
func (m *userStorage) SetUserState(
	userID string, state teak.UserState) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C("users").Update(
		bson.M{
			"id": userID,
		},
		bson.M{
			"$set": bson.M{
				"state": state,
			},
		})
	return teak.LogError("UMan:Mongo", err)
}

//VerifyUser - sets state of an user account to verified based on userID
//and verification ID
func (m *userStorage) VerifyUser(userID, verID string) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C("users").Update(
		bson.M{
			"$and": []bson.M{
				bson.M{"id": userID},
				bson.M{"verID": verID},
			},
		},
		bson.M{
			"$set": bson.M{
				"state":      teak.Active,
				"verifiedAt": time.Now(),
				"verID":      "",
			},
		})
	return teak.LogError("UMan:Mongo", err)
}

//CreateIndices - creates mongoDB indeces for tables used for user management
func (m *userStorage) CreateIndices() (err error) {
	conn := DefaultConn()
	defer conn.Close()
	err = conn.C("users").EnsureIndex(mgo.Index{
		Key:        []string{"id", "email"},
		Unique:     true,
		DropDups:   true,
		Background: true, // See notes.
		Sparse:     true,
	})
	return err
}

//CleanData - cleans user management related data from database
func (m *userStorage) CleanData() (err error) {
	conn := DefaultConn()
	defer conn.Close()
	_, err = conn.C("users").RemoveAll(bson.M{})
	return err
}

//UpdateProfile - updates user details - this should be used when user logged in
//is updating own user account
func (m *userStorage) UpdateProfile(user *teak.User) (err error) {
	conn := DefaultConn()
	defer conn.Close()
	user.FullName = user.FirstName + " " + user.LastName
	err = conn.C("users").Update(
		bson.M{
			"id": user.ID,
		}, bson.M{
			"$set": bson.M{
				"email":      user.Email,
				"firstName":  user.FirstName,
				"lastName":   user.LastName,
				"title":      user.Title,
				"fullName":   user.FullName,
				"modifiedAt": time.Now(),
				"modifiedBy": user.FullName,
			},
		},
	)
	return teak.LogError("UMan:Mongo", err)
}
