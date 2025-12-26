package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseModel defines the common fields for all models
type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// SoftDeletable interface for models that support soft delete
type SoftDeletable interface {
	GetDeletedAt() *time.Time
	SetDeletedAt(t *time.Time)
}

// GetDeletedAt returns the deleted_at timestamp
func (bm *BaseModel) GetDeletedAt() *time.Time {
	return bm.DeletedAt
}

// SetDeletedAt sets the deleted_at timestamp
func (bm *BaseModel) SetDeletedAt(t *time.Time) {
	bm.DeletedAt = t
}

// BaseRepository provides generic CRUD operations for MongoDB collections
type BaseRepository struct {
	collection   *mongo.Collection
	client       *Client
	softDelete   bool
	queryBuilder *QueryBuilder
}

// RepositoryConfig holds configuration for creating a repository
type RepositoryConfig struct {
	Collection *mongo.Collection
	Client     *Client
	SoftDelete bool // Enable soft delete functionality
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(config RepositoryConfig) *BaseRepository {
	return &BaseRepository{
		collection:   config.Collection,
		client:       config.Client,
		softDelete:   config.SoftDelete,
		queryBuilder: NewQueryBuilder(),
	}
}

// GetCollection returns the underlying collection
func (r *BaseRepository) GetCollection() *mongo.Collection {
	return r.collection
}

// GetQueryBuilder returns a new query builder instance
func (r *BaseRepository) GetQueryBuilder() *QueryBuilder {
	return NewQueryBuilder()
}

// addSoftDeleteFilter adds soft delete filter if enabled
func (r *BaseRepository) addSoftDeleteFilter(filter bson.M) bson.M {
	if r.softDelete {
		if filter == nil {
			filter = bson.M{}
		}
		filter["deleted_at"] = bson.M{"$exists": false}
	}
	return filter
}

// Create inserts a new document
func (r *BaseRepository) Create(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	// Set timestamps if document has BaseModel
	if model, ok := document.(interface {
		SetCreatedAt(time.Time)
		SetUpdatedAt(time.Time)
	}); ok {
		now := time.Now()
		model.SetCreatedAt(now)
		model.SetUpdatedAt(now)
	}

	// Handle bson.M
	if doc, ok := document.(bson.M); ok {
		now := time.Now()
		doc["created_at"] = now
		doc["updated_at"] = now
	}

	return r.collection.InsertOne(ctx, document)
}

// CreateMany inserts multiple documents
func (r *BaseRepository) CreateMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
	now := time.Now()
	for _, document := range documents {
		if model, ok := document.(interface {
			SetCreatedAt(time.Time)
			SetUpdatedAt(time.Time)
		}); ok {
			model.SetCreatedAt(now)
			model.SetUpdatedAt(now)
		}

		if doc, ok := document.(bson.M); ok {
			doc["created_at"] = now
			doc["updated_at"] = now
		}
	}

	return r.collection.InsertMany(ctx, documents)
}

// FindByID finds a document by ID
func (r *BaseRepository) FindByID(ctx context.Context, id primitive.ObjectID, result interface{}) error {
	filter := bson.M{"_id": id}
	filter = r.addSoftDeleteFilter(filter)
	return r.collection.FindOne(ctx, filter).Decode(result)
}

// FindOne finds a single document matching the filter
func (r *BaseRepository) FindOne(ctx context.Context, filter bson.M, result interface{}, opts ...*options.FindOneOptions) error {
	filter = r.addSoftDeleteFilter(filter)
	return r.collection.FindOne(ctx, filter, opts...).Decode(result)
}

// Find finds multiple documents matching the filter
func (r *BaseRepository) Find(ctx context.Context, filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	filter = r.addSoftDeleteFilter(filter)
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// FindAll returns all documents (with soft delete filter if enabled)
func (r *BaseRepository) FindAll(ctx context.Context, results interface{}, opts ...*options.FindOptions) error {
	return r.Find(ctx, bson.M{}, results, opts...)
}

// Update updates a single document matching the filter
func (r *BaseRepository) Update(ctx context.Context, filter bson.M, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = r.addSoftDeleteFilter(filter)

	// Add updated_at timestamp
	if updateDoc, ok := update.(bson.M); ok {
		if set, exists := updateDoc["$set"]; exists {
			if setDoc, ok := set.(bson.M); ok {
				setDoc["updated_at"] = time.Now()
			}
		} else {
			updateDoc["$set"] = bson.M{"updated_at": time.Now()}
		}
	}

	return r.collection.UpdateOne(ctx, filter, update, opts...)
}

// UpdateByID updates a document by ID
func (r *BaseRepository) UpdateByID(ctx context.Context, id primitive.ObjectID, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return r.Update(ctx, bson.M{"_id": id}, update, opts...)
}

// UpdateMany updates multiple documents matching the filter
func (r *BaseRepository) UpdateMany(ctx context.Context, filter bson.M, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = r.addSoftDeleteFilter(filter)

	// Add updated_at timestamp
	if updateDoc, ok := update.(bson.M); ok {
		if set, exists := updateDoc["$set"]; exists {
			if setDoc, ok := set.(bson.M); ok {
				setDoc["updated_at"] = time.Now()
			}
		} else {
			updateDoc["$set"] = bson.M{"updated_at": time.Now()}
		}
	}

	return r.collection.UpdateMany(ctx, filter, update, opts...)
}

// Delete deletes a single document (soft delete if enabled)
func (r *BaseRepository) Delete(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	filter = r.addSoftDeleteFilter(filter)

	if r.softDelete {
		// Soft delete: set deleted_at timestamp
		now := time.Now()
		update := bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
			},
		}
		result, err := r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, err
		}
		return &mongo.DeleteResult{DeletedCount: result.ModifiedCount}, nil
	}

	// Hard delete
	return r.collection.DeleteOne(ctx, filter, opts...)
}

// DeleteByID deletes a document by ID (soft delete if enabled)
func (r *BaseRepository) DeleteByID(ctx context.Context, id primitive.ObjectID, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return r.Delete(ctx, bson.M{"_id": id}, opts...)
}

// DeleteMany deletes multiple documents (soft delete if enabled)
func (r *BaseRepository) DeleteMany(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	filter = r.addSoftDeleteFilter(filter)

	if r.softDelete {
		// Soft delete: set deleted_at timestamp
		now := time.Now()
		update := bson.M{
			"$set": bson.M{
				"deleted_at": now,
				"updated_at": now,
			},
		}
		result, err := r.collection.UpdateMany(ctx, filter, update)
		if err != nil {
			return nil, err
		}
		return &mongo.DeleteResult{DeletedCount: result.ModifiedCount}, nil
	}

	// Hard delete
	return r.collection.DeleteMany(ctx, filter, opts...)
}

// HardDelete permanently deletes a document (bypasses soft delete)
func (r *BaseRepository) HardDelete(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	// Remove soft delete filter to allow deleting soft-deleted documents
	return r.collection.DeleteOne(ctx, filter, opts...)
}

// HardDeleteByID permanently deletes a document by ID
func (r *BaseRepository) HardDeleteByID(ctx context.Context, id primitive.ObjectID, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return r.HardDelete(ctx, bson.M{"_id": id}, opts...)
}

// Restore restores a soft-deleted document
func (r *BaseRepository) Restore(ctx context.Context, filter bson.M) (*mongo.UpdateResult, error) {
	if !r.softDelete {
		return nil, errors.New("soft delete is not enabled for this repository")
	}

	// Remove soft delete filter and add deleted_at exists check
	filter["deleted_at"] = bson.M{"$exists": true}

	update := bson.M{
		"$unset": bson.M{"deleted_at": ""},
		"$set":   bson.M{"updated_at": time.Now()},
	}

	return r.collection.UpdateOne(ctx, filter, update)
}

// RestoreByID restores a soft-deleted document by ID
func (r *BaseRepository) RestoreByID(ctx context.Context, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	return r.Restore(ctx, bson.M{"_id": id})
}

// Count counts documents matching the filter
func (r *BaseRepository) Count(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int64, error) {
	filter = r.addSoftDeleteFilter(filter)
	return r.collection.CountDocuments(ctx, filter, opts...)
}

// Exists checks if a document matching the filter exists
func (r *BaseRepository) Exists(ctx context.Context, filter bson.M) (bool, error) {
	filter = r.addSoftDeleteFilter(filter)
	count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Paginate performs pagination with the repository's soft delete filter
func (r *BaseRepository) Paginate(ctx context.Context, filter bson.M, params *PaginationParams, results interface{}) (*PaginationResult, error) {
	filter = r.addSoftDeleteFilter(filter)
	return Paginate(ctx, r.collection, filter, params, results)
}

// Aggregate performs aggregation with soft delete filter prepended
func (r *BaseRepository) Aggregate(ctx context.Context, pipeline []bson.M, results interface{}, opts ...*options.AggregateOptions) error {
	if r.softDelete {
		// Prepend soft delete filter
		pipeline = append([]bson.M{
			{"$match": bson.M{"deleted_at": bson.M{"$exists": false}}},
		}, pipeline...)
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// Transaction executes a function within a transaction
func (r *BaseRepository) Transaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	if r.client == nil {
		return errors.New("client is required for transaction support")
	}
	return r.client.Transaction(ctx, fn)
}

// TransactionWithOptions executes a function within a transaction with custom options
func (r *BaseRepository) TransactionWithOptions(ctx context.Context, fn func(sessCtx mongo.SessionContext) error, opts *options.TransactionOptions) error {
	if r.client == nil {
		return errors.New("client is required for transaction support")
	}
	return r.client.TransactionWithOptions(ctx, fn, opts)
}

// Index management functions

// IndexConfig defines configuration for creating an index
type IndexConfig struct {
	Keys    bson.D
	Options *options.IndexOptions
}

// CreateIndex creates a single index
func (r *BaseRepository) CreateIndex(ctx context.Context, keys bson.D, opts ...*options.IndexOptions) (string, error) {
	indexModel := mongo.IndexModel{
		Keys: keys,
	}
	if len(opts) > 0 {
		indexModel.Options = opts[0]
	}
	return r.collection.Indexes().CreateOne(ctx, indexModel)
}

// CreateIndexes creates multiple indexes
func (r *BaseRepository) CreateIndexes(ctx context.Context, configs []IndexConfig) ([]string, error) {
	models := make([]mongo.IndexModel, len(configs))
	for i, config := range configs {
		models[i] = mongo.IndexModel{
			Keys:    config.Keys,
			Options: config.Options,
		}
	}
	return r.collection.Indexes().CreateMany(ctx, models)
}

// CreateUniqueIndex creates a unique index
func (r *BaseRepository) CreateUniqueIndex(ctx context.Context, keys bson.D, name string) (string, error) {
	opts := options.Index().SetUnique(true)
	if name != "" {
		opts.SetName(name)
	}
	return r.CreateIndex(ctx, keys, opts)
}

// CreateTextIndex creates a text index for full-text search
func (r *BaseRepository) CreateTextIndex(ctx context.Context, fields ...string) (string, error) {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}
	return r.CreateIndex(ctx, keys)
}

// CreateTTLIndex creates a TTL (Time To Live) index for automatic document expiration
func (r *BaseRepository) CreateTTLIndex(ctx context.Context, field string, expireAfterSeconds int32) (string, error) {
	keys := bson.D{{Key: field, Value: 1}}
	opts := options.Index().SetExpireAfterSeconds(expireAfterSeconds)
	return r.CreateIndex(ctx, keys, opts)
}

// CreateCompoundIndex creates a compound index on multiple fields
func (r *BaseRepository) CreateCompoundIndex(ctx context.Context, fields map[string]int, unique bool) (string, error) {
	keys := bson.D{}
	for field, order := range fields {
		keys = append(keys, bson.E{Key: field, Value: order})
	}

	var opts *options.IndexOptions
	if unique {
		opts = options.Index().SetUnique(true)
	}

	return r.CreateIndex(ctx, keys, opts)
}

// DropIndex drops an index by name
func (r *BaseRepository) DropIndex(ctx context.Context, name string) error {
	_, err := r.collection.Indexes().DropOne(ctx, name)
	return err
}

// DropAllIndexes drops all indexes except the default _id index
func (r *BaseRepository) DropAllIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().DropAll(ctx)
	return err
}

// ListIndexes lists all indexes on the collection
func (r *BaseRepository) ListIndexes(ctx context.Context) ([]bson.M, error) {
	cursor, err := r.collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return nil, err
	}
	return indexes, nil
}

// EnsureIndexes ensures that specified indexes exist, creating them if necessary
func (r *BaseRepository) EnsureIndexes(ctx context.Context, configs []IndexConfig) error {
	_, err := r.CreateIndexes(ctx, configs)
	if err != nil {
		// Check if error is due to index already existing
		if mongo.IsDuplicateKeyError(err) ||
			(err != nil && (err.Error() == "index already exists" ||
				err.Error() == "IndexOptionsConflict")) {
			return nil
		}
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}
	return nil
}

// Helper methods for timestamp management

// SetCreatedAt sets the created_at timestamp on BaseModel
func (bm *BaseModel) SetCreatedAt(t time.Time) {
	bm.CreatedAt = t
}

// SetUpdatedAt sets the updated_at timestamp on BaseModel
func (bm *BaseModel) SetUpdatedAt(t time.Time) {
	bm.UpdatedAt = t
}
