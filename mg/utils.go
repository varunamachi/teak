package mg

import (
	"context"

	"github.com/varunamachi/teak"
	"go.mongodb.org/mongo-driver/mongo"
)

// Decode - convenience method for decoding mongo.SingleResult
func Decode(res *mongo.SingleResult, out interface{}) error {
	if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(out)
}

// ReadAllAndClose - reads all data from the cursor and closes it
func ReadAllAndClose(
	gtx context.Context,
	cur *mongo.Cursor,
	out interface{}) error {

	defer func() {
		if err := cur.Close(gtx); err != nil {
			teak.LogErrorX("teak.mongo", "Failed to close cursor", err)
		}
	}()
	return cur.All(gtx, out)
}

// ReadOneAndClose - reads one data item from the cursor and closes it
func ReadOneAndClose(
	gtx context.Context,
	cur *mongo.Cursor,
	out interface{}) error {

	defer func() {
		if err := cur.Close(gtx); err != nil {
			teak.LogErrorX("teak.mongo", "Failed to close cursor", err)
		}
	}()
	return cur.Decode(out)
}
