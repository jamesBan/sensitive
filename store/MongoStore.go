package github.com/jamesBan/sensitive/store

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type MongoStore struct {
	DSN             string
	DB              string
	Collection      string
	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
	mongoVersion    uint64
	locker          sync.RWMutex
}

func NewMongoStore(dsn, db, collection string, timeout time.Duration) (*MongoStore, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	coll := client.Database(db).Collection(collection)

	manage := &MongoStore{
		DSN:             dsn,
		DB:              db,
		Collection:      collection,
		mongoClient:     client,
		mongoCollection: coll,
	}

	return manage, nil
}

func (s *MongoStore) Write(word string) error {
	if len(word) < 1 {
		return errors.New("empty word")
	}

	if s.exists(word) {
		return errors.Errorf("word %s exists", word)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err := s.mongoCollection.InsertOne(ctx, bson.M{"value": word})
	if err != nil {
		return err
	}

	s.locker.Lock()
	defer s.locker.Unlock()
	s.mongoVersion++

	return nil
}

func (s *MongoStore) exists(word string) bool {
	var wordResult bson.M
	err := s.mongoCollection.FindOne(nil, bson.M{"value": word}).Decode(&wordResult)
	if err != nil {
		return false
	}

	return true
}

func (s *MongoStore) Remove(word string) error {
	if len(word) < 1 {
		return errors.New("empty word")
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	_, err := s.mongoCollection.DeleteOne(ctx, bson.M{"value": word})
	if err != nil {
		return err
	}

	s.locker.Lock()
	defer s.locker.Unlock()
	s.mongoVersion++

	return nil
}

func (s *MongoStore) Version() uint64 {
	return s.mongoVersion
}

func (s *MongoStore) ReadAll() <-chan string {
	resultChannel := make(chan string)

	go func() {
		cur, err := s.mongoCollection.Find(nil, bson.M{})
		if err != nil {
			panic(err)
		}
		defer cur.Close(nil)
		defer close(resultChannel)

		for cur.Next(nil) {
			var result bson.M
			err := cur.Decode(&result)
			if err != nil {
				continue
			}

			resultChannel <- result["value"].(string)
		}

	}()

	return resultChannel
}
