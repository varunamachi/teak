package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/varunamachi/teak"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/hlandau/passlib.v1"
	"gopkg.in/mgo.v2"
)

//userStorage - mongodb storage for user information
type userStorage struct{}

//NewUserStorage - creates a new user storage based on mongodb
func NewUserStorage() teak.UserStorage {
	return &userStorage{}
}

//CreateUser - creates user in database
func (m *userStorage) CreateUser(
	gtx context.Context,
	user *teak.User) (string, error) {
	if err := teak.UpdateUserInfo(user); err != nil {
		err = teak.LogErrorX("t.user.mongo",
			"Failed to create user, user storage not properly configured", err)
		return "", err
	}
	if err := m.validateForSuper(gtx, user.Auth); err != nil {
		return "", err
	}
	_, err := C("users").InsertOne(gtx, user)
	return user.ID, teak.LogError("t.user.mongo", err)
}

//UpdateUser - updates user in database
func (m *userStorage) UpdateUser(
	gtx context.Context, user *teak.User) error {
	if err := m.validateForSuper(gtx, user.Auth); err != nil {
		return err
	}
	_, err := C("users").UpdateOne(gtx, bson.M{"id": user.ID}, user)
	return teak.LogError("t.user.mongo", err)
}

//DeleteUser - deletes user with given user ID
func (m *userStorage) DeleteUser(
	gtx context.Context, userID string) error {
	_, err := C("users").DeleteOne(gtx, bson.M{"id": userID})
	return teak.LogError("t.user.mongo", err)
}

//GetUser - gets details of the user corresponding to ID
func (m *userStorage) GetUser(gtx context.Context,
	userID string) (*teak.User, error) {
	user := &teak.User{}
	res := C("users").FindOne(gtx, bson.M{"id": userID})
	if err := res.Decode(user); err != nil {
		return nil, teak.LogError("t.user.mongo", err)
	}
	return user, nil
}

//GetUsers - gets all users based on offset, limit and filter
func (m *userStorage) GetUsers(
	gtx context.Context,
	offset, limit int64,
	filter *teak.Filter) ([]*teak.User, error) {
	selector := generateSelector(filter)
	users := make([]*teak.User, 0, limit)
	fopts := options.Find().SetSkip(offset).SetLimit(limit)
	cur, err := C("users").Find(gtx, selector, fopts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(gtx)
	err = cur.All(gtx, &users)
	return users, teak.LogError("t.user.mongo", err)
}

//GetCount - gives the number of user selected by given filter
func (m *userStorage) GetCount(
	gtx context.Context, filter *teak.Filter) (int64, error) {
	selector := generateSelector(filter)
	count, err := C("users").CountDocuments(gtx, selector)
	return count, teak.LogError("t.user.mongo", err)
}

//GetUsersWithCount - Get users with total count
func (m *userStorage) GetUsersWithCount(
	gtx context.Context,
	offset, limit int64,
	filter *teak.Filter) (int64, []*teak.User, error) {
	var selector bson.M
	selector = generateSelector(filter)
	users := make([]*teak.User, 0, limit)

	fopts := options.Find().
		SetSkip(offset).
		SetLimit(limit).
		SetSort(GetSort("-created"))
	cur, err := C("users").Find(gtx, selector, fopts)
	if err != nil {
		return 0, nil, teak.LogError("t.user.mongo", err)
	}
	err = cur.All(gtx, &users)
	if err != nil {
		return 0, nil, teak.LogError("t.user.mongo", err)
	}

	count, err := C("users").CountDocuments(gtx, selector)
	if err != nil {
		return 0, nil, teak.LogError("t.user.mongo", err)
	}
	return count, users, nil
}

//ResetPassword - sets password of a unauthenticated user
func (m *userStorage) ResetPassword(
	gtx context.Context, userID, oldPwd, newPwd string) error {
	var newHash string
	newHash, err := passlib.Hash(newPwd)
	if err != nil {
		return err
	}
	if err = m.ValidateUser(gtx, userID, oldPwd); err != nil {
		return errors.New("Could not match old password")
	}
	_, err = C("secret").UpdateOne(
		gtx,
		bson.M{
			"userID": userID,
		},
		bson.M{
			"$set": bson.M{
				"phash": newHash,
			},
		},
	)
	return teak.LogError("t.user.mongo", err)
}

//SetPassword - sets password of a already authenticated user, old password
//is not required
func (m *userStorage) SetPassword(
	gtx context.Context, userID, newPwd string) (err error) {
	var newHash string
	newHash, err = passlib.Hash(newPwd)
	if err != nil {
		return teak.LogError("t.user.mongo", err)
	}
	m.setPasswordHash(gtx, userID, newHash)
	return teak.LogError("t.user.mongo", err)
}

func (m *userStorage) setPasswordHash(
	gtx context.Context, userID, hash string) (err error) {
	// _, err = C("secret").Upsert(
	// 	bson.M{
	// 		"userID": userID,
	// 	},
	// 	bson.M{
	// 		"$set": bson.M{
	// 			"userID": userID,
	// 			"phash":  hash,
	// 		},
	// 	})
	// return err
	return nil
}

//ValidateUser - validates user ID and password
func (m *userStorage) ValidateUser(
	gtx context.Context, userID, password string) error {

	secret := bson.M{}
	fopts := options.Find().SetProjection(bson.M{"phash": 1, "_id": 0})
	res := C("secret").FindOne(gtx, bson.M{"userID": userID}, fopts)

	if err := res.Err(); err != nil {
		return teak.LogError("t.user.mongo", err)
	}

	err := res.Decode(&secret)
	if err != nil {
		return teak.LogError("t.user.mongo", err)
	}

	storedHash, ok := secret["phash"].(string)
	if !ok {
		return errors.New("Failed to varify password")
	}

	newHash, err := passlib.Verify(password, storedHash)
	if err != nil {
		return teak.LogError("t.user.mongo", err)
	}

	if newHash != "" {
		err = m.setPasswordHash(gtx, userID, newHash)
	}
	return teak.LogError("t.user.mongo", err)
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
	gtx context.Context,
	alevel teak.AuthLevel) error {
	if alevel != teak.Super {
		return nil
	}
	count, err := C("users").CountDocuments(gtx, bson.M{"auth": 0})
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
				{"id": userID},
				{"verID": verID},
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
