package middleware

import (
	"net/http"
	"time"

	"github.com/gorilla/context"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	Session *mgo.Session
	DB      string
}

func NewDB(addrs []string, db, user, pwd string, log *Logger) *MongoDB {
	mdb := MongoDB{
		DB: db,
	}
	info := &mgo.DialInfo{
		Addrs:    addrs,
		Timeout:  10 * time.Second,
		Database: db,
		Username: user,
		Password: pwd,
	}
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatalf("failed to connect to mongodb at %s -- %s", addrs, err)
	}
	mdb.Session = session

	return &mdb
}

func (mdb *MongoDB) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	reqSession := mdb.Session.Clone()
	defer reqSession.Close()
	db := reqSession.DB(mdb.DB)
	context.Set(req, mdb.DB, db)
	next(rw, req)
}

func (mdb *MongoDB) GetDB(req *http.Request) *mgo.Database {
	if reqv := context.Get(req, mdb.DB); reqv != nil {
		return reqv.(*mgo.Database)
	}
	return nil
}

func (mdb *MongoDB) Close() {
	mdb.Session.Close()
}
