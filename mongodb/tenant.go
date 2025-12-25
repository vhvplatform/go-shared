package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TenantAware is an interface for models that support multi-tenancy
type TenantAware interface {
	GetTenantID() string
	SetTenantID(tenantID string)
}

// TenantRepository provides tenant-scoped database operations
type TenantRepository struct {
	collection *mongo.Collection
	tenantID   string
}

// NewTenantRepository creates a new tenant-scoped repository
func NewTenantRepository(collection *mongo.Collection, tenantID string) *TenantRepository {
	return &TenantRepository{
		collection: collection,
		tenantID:   tenantID,
	}
}

// addTenantFilter adds tenant_id to the filter
func (tr *TenantRepository) addTenantFilter(filter bson.M) bson.M {
	if filter == nil {
		filter = bson.M{}
	}
	filter["tenant_id"] = tr.tenantID
	return filter
}

// FindOne finds a single document with tenant isolation
func (tr *TenantRepository) FindOne(ctx context.Context, filter bson.M, result interface{}, opts ...*options.FindOneOptions) error {
	filter = tr.addTenantFilter(filter)
	return tr.collection.FindOne(ctx, filter, opts...).Decode(result)
}

// Find finds multiple documents with tenant isolation
func (tr *TenantRepository) Find(ctx context.Context, filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	filter = tr.addTenantFilter(filter)
	cursor, err := tr.collection.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// InsertOne inserts a single document with tenant_id
func (tr *TenantRepository) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	// Set tenant_id if document implements TenantAware
	if tenantAware, ok := document.(TenantAware); ok {
		tenantAware.SetTenantID(tr.tenantID)
	}

	// If document is a map or bson.M, add tenant_id
	switch doc := document.(type) {
	case bson.M:
		doc["tenant_id"] = tr.tenantID
	case map[string]interface{}:
		doc["tenant_id"] = tr.tenantID
	}

	return tr.collection.InsertOne(ctx, document, opts...)
}

// InsertMany inserts multiple documents with tenant_id
func (tr *TenantRepository) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	// Set tenant_id for each document
	for _, document := range documents {
		if tenantAware, ok := document.(TenantAware); ok {
			tenantAware.SetTenantID(tr.tenantID)
		}

		switch doc := document.(type) {
		case bson.M:
			doc["tenant_id"] = tr.tenantID
		case map[string]interface{}:
			doc["tenant_id"] = tr.tenantID
		}
	}

	return tr.collection.InsertMany(ctx, documents, opts...)
}

// UpdateOne updates a single document with tenant isolation
func (tr *TenantRepository) UpdateOne(ctx context.Context, filter bson.M, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = tr.addTenantFilter(filter)
	return tr.collection.UpdateOne(ctx, filter, update, opts...)
}

// UpdateMany updates multiple documents with tenant isolation
func (tr *TenantRepository) UpdateMany(ctx context.Context, filter bson.M, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = tr.addTenantFilter(filter)
	return tr.collection.UpdateMany(ctx, filter, update, opts...)
}

// DeleteOne deletes a single document with tenant isolation
func (tr *TenantRepository) DeleteOne(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	filter = tr.addTenantFilter(filter)
	return tr.collection.DeleteOne(ctx, filter, opts...)
}

// DeleteMany deletes multiple documents with tenant isolation
func (tr *TenantRepository) DeleteMany(ctx context.Context, filter bson.M, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	filter = tr.addTenantFilter(filter)
	return tr.collection.DeleteMany(ctx, filter, opts...)
}

// CountDocuments counts documents with tenant isolation
func (tr *TenantRepository) CountDocuments(ctx context.Context, filter bson.M, opts ...*options.CountOptions) (int64, error) {
	filter = tr.addTenantFilter(filter)
	return tr.collection.CountDocuments(ctx, filter, opts...)
}

// Aggregate performs aggregation with tenant filter
func (tr *TenantRepository) Aggregate(ctx context.Context, pipeline []bson.M, results interface{}, opts ...*options.AggregateOptions) error {
	// Prepend a $match stage with tenant_id - pre-allocate for efficiency
	tenantMatch := bson.M{"$match": bson.M{"tenant_id": tr.tenantID}}
	tenantPipeline := make([]bson.M, 0, len(pipeline)+1)
	tenantPipeline = append(tenantPipeline, tenantMatch)
	tenantPipeline = append(tenantPipeline, pipeline...)

	cursor, err := tr.collection.Aggregate(ctx, tenantPipeline, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// Paginate performs tenant-aware pagination
func (tr *TenantRepository) Paginate(ctx context.Context, filter bson.M, params *PaginationParams, results interface{}) (*PaginationResult, error) {
	filter = tr.addTenantFilter(filter)
	return Paginate(ctx, tr.collection, filter, params, results)
}

// EnsureTenantIndex creates an index on tenant_id field
func (tr *TenantRepository) EnsureTenantIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "tenant_id", Value: 1}},
	}
	_, err := tr.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create tenant_id index: %w", err)
	}
	return nil
}

// EnsureCompoundIndex creates a compound index including tenant_id
func (tr *TenantRepository) EnsureCompoundIndex(ctx context.Context, fields ...string) error {
	keys := bson.D{{Key: "tenant_id", Value: 1}}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}

	indexModel := mongo.IndexModel{
		Keys: keys,
	}
	_, err := tr.collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create compound index: %w", err)
	}
	return nil
}

// GetCollection returns the underlying collection
func (tr *TenantRepository) GetCollection() *mongo.Collection {
	return tr.collection
}

// GetTenantID returns the tenant ID
func (tr *TenantRepository) GetTenantID() string {
	return tr.tenantID
}
