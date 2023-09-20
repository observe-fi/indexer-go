package db

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (p *Provider) LoadCollection(name string) *Collection {
	col, ok := p.collections[name]
	if ok {
		return col
	}
	coll := p.db.Collection(name)
	col = &Collection{col: coll, provider: p}
	p.collections[name] = col
	return col
}

func (c *Collection) Create(document interface{}) error {
	_, err := c.col.InsertOne(context.Background(), document)
	return err
}

func (c *Collection) Upsert(filter interface{}, document interface{}) error {
	opts := options.Update().SetUpsert(true)
	_, err := c.col.UpdateOne(context.Background(), filter, document, opts)
	return err
}

// ReadOne retrieves a single document from the MongoDB collection.
func (c *Collection) ReadOne(ctx context.Context, filter interface{}, result interface{}) error {
	return c.col.FindOne(ctx, filter).Decode(result)
}

// ReadID retrieves a single document from the MongoDB collection.
func (c *Collection) ReadID(ctx context.Context, key string, result interface{}) error {
	return c.col.FindOne(ctx, LookupID(key)).Decode(result)
}

// ReadAll retrieves multiple documents from the MongoDB collection.
func (c *Collection) ReadAll(ctx context.Context, filter interface{}, results interface{}) error {
	cur, err := c.col.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, results); err != nil {
		return err
	}

	return nil
}

// Update updates a document in the MongoDB collection.
func (c *Collection) Update(ctx context.Context, filter interface{}, update interface{}) error {
	updateResult, err := c.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if updateResult.MatchedCount == 0 {
		return errors.New("no document found to update")
	}

	return nil
}

// Delete removes a document from the MongoDB collection.
func (c *Collection) Delete(ctx context.Context, filter interface{}) error {
	deleteResult, err := c.col.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if deleteResult.DeletedCount == 0 {
		return errors.New("no document found to delete")
	}

	return nil
}
