package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	databasepb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	instancepb "cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/iterator"

	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/get_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/list_products"
	"github.com/tshubham2/catalog-proj/internal/app/product/repo"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/activate_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/apply_discount"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/create_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/update_product"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
)

const (
	projectID  = "test-project"
	instanceID = "test-instance"
)

var (
	spannerClient     *spanner.Client
	createProductUC   *create_product.Interactor
	updateProductUC   *update_product.Interactor
	applyDiscountUC   *apply_discount.ApplyInteractor
	removeDiscountUC  *apply_discount.RemoveInteractor
	activateUC        *activate_product.ActivateInteractor
	deactivateUC      *activate_product.DeactivateInteractor
	getProductQuery   *get_product.Handler
	listProductsQuery *list_products.Handler
	testClock         clock.Clock
)

func TestMain(m *testing.M) {
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	databaseID := fmt.Sprintf("testdb_%d", time.Now().UnixNano())
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)

	if err := provisionEmulator(ctx, databaseID); err != nil {
		fmt.Fprintf(os.Stderr, "skipping E2E tests (emulator not available): %v\n", err)
		os.Exit(0)
	}

	client, err := spanner.NewClient(ctx, dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create spanner client: %v\n", err)
		os.Exit(1)
	}
	spannerClient = client

	wireUsecases(client)

	code := m.Run()
	client.Close()
	os.Exit(code)
}

// --- Test flows ---

func TestProductCreationFlow(t *testing.T) {
	ctx := context.Background()

	productID, err := createProductUC.Execute(ctx, create_product.Request{
		Name:        "Widget Pro",
		Description: "Premium widget",
		Category:    "gadgets",
		BasePrice:   big.NewRat(1999, 100),
	})
	require.NoError(t, err)
	require.NotEmpty(t, productID)

	product, err := getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "Widget Pro", product.Name)
	assert.Equal(t, "gadgets", product.Category)
	assert.Equal(t, "19.99", product.BasePrice)
	assert.Equal(t, "19.99", product.EffectivePrice)
	assert.Equal(t, "active", product.Status)

	events := getOutboxEvents(t, ctx, productID)
	require.Len(t, events, 1)
	assert.Equal(t, "product.created", events[0].eventType)
}

func TestProductUpdateFlow(t *testing.T) {
	ctx := context.Background()

	productID := createTestProduct(t, ctx, "Original Name", "electronics")

	newName := "Updated Name"
	newDesc := "Updated description"
	err := updateProductUC.Execute(ctx, update_product.Request{
		ProductID:   productID,
		Name:        &newName,
		Description: &newDesc,
	})
	require.NoError(t, err)

	product, err := getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Updated description", product.Description)
	assert.Equal(t, "electronics", product.Category) // unchanged
}

func TestDiscountApplicationFlow(t *testing.T) {
	ctx := context.Background()

	productID := createTestProduct(t, ctx, "Discounted Item", "clothing")

	now := time.Now().UTC()
	err := applyDiscountUC.Execute(ctx, apply_discount.ApplyRequest{
		ProductID:  productID,
		Percentage: big.NewRat(25, 1), // 25%
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(24 * time.Hour),
	})
	require.NoError(t, err)

	product, err := getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "49.99", product.BasePrice)
	assert.Equal(t, "37.49", product.EffectivePrice) // 49.99 * 0.75
	require.NotNil(t, product.DiscountPercent)
	assert.Equal(t, "25.00", *product.DiscountPercent)

	events := getOutboxEvents(t, ctx, productID)
	eventTypes := make([]string, len(events))
	for i, e := range events {
		eventTypes[i] = e.eventType
	}
	assert.Contains(t, eventTypes, "discount.applied")
}

func TestRemoveDiscountFlow(t *testing.T) {
	ctx := context.Background()

	productID := createTestProduct(t, ctx, "Temp Discount Item", "clothing")

	now := time.Now().UTC()
	require.NoError(t, applyDiscountUC.Execute(ctx, apply_discount.ApplyRequest{
		ProductID:  productID,
		Percentage: big.NewRat(10, 1),
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(24 * time.Hour),
	}))

	err := removeDiscountUC.Execute(ctx, apply_discount.RemoveRequest{ProductID: productID})
	require.NoError(t, err)

	product, err := getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, product.BasePrice, product.EffectivePrice)
	assert.Nil(t, product.DiscountPercent)
}

func TestProductActivationDeactivation(t *testing.T) {
	ctx := context.Background()

	productID := createTestProduct(t, ctx, "Toggle Item", "electronics")

	// Deactivate
	err := deactivateUC.Execute(ctx, activate_product.Request{ProductID: productID})
	require.NoError(t, err)

	product, err := getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "inactive", product.Status)

	// Re-activate
	err = activateUC.Execute(ctx, activate_product.Request{ProductID: productID})
	require.NoError(t, err)

	product, err = getProductQuery.Execute(ctx, productID)
	require.NoError(t, err)
	assert.Equal(t, "active", product.Status)
}

func TestBusinessRuleValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("cannot apply discount to inactive product", func(t *testing.T) {
		productID := createTestProduct(t, ctx, "Inactive Item", "electronics")
		require.NoError(t, deactivateUC.Execute(ctx, activate_product.Request{ProductID: productID}))

		now := time.Now().UTC()
		err := applyDiscountUC.Execute(ctx, apply_discount.ApplyRequest{
			ProductID:  productID,
			Percentage: big.NewRat(10, 1),
			StartDate:  now.Add(-time.Hour),
			EndDate:    now.Add(24 * time.Hour),
		})
		assert.ErrorIs(t, err, domain.ErrProductNotActive)
	})

	t.Run("cannot activate already active product", func(t *testing.T) {
		productID := createTestProduct(t, ctx, "Already Active", "electronics")
		err := activateUC.Execute(ctx, activate_product.Request{ProductID: productID})
		assert.ErrorIs(t, err, domain.ErrProductAlreadyActive)
	})

	t.Run("cannot remove discount when none exists", func(t *testing.T) {
		productID := createTestProduct(t, ctx, "No Discount", "electronics")
		err := removeDiscountUC.Execute(ctx, apply_discount.RemoveRequest{ProductID: productID})
		assert.ErrorIs(t, err, domain.ErrNoActiveDiscount)
	})
}

func TestListActiveProducts(t *testing.T) {
	ctx := context.Background()

	category := fmt.Sprintf("list-test-%d", time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		createTestProductWithCategory(t, ctx, fmt.Sprintf("Item %d", i), category)
	}

	result, err := listProductsQuery.Execute(ctx, list_products.Params{
		PageSize: 10,
		Category: category,
	})
	require.NoError(t, err)
	assert.Len(t, result.Products, 3)
	assert.Empty(t, result.NextPageToken)
}

func TestListActiveProducts_Pagination(t *testing.T) {
	ctx := context.Background()

	category := fmt.Sprintf("page-test-%d", time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		createTestProductWithCategory(t, ctx, fmt.Sprintf("Page Item %d", i), category)
	}

	// First page
	result, err := listProductsQuery.Execute(ctx, list_products.Params{
		PageSize: 2,
		Category: category,
	})
	require.NoError(t, err)
	assert.Len(t, result.Products, 2)
	assert.NotEmpty(t, result.NextPageToken)

	// Second page
	result2, err := listProductsQuery.Execute(ctx, list_products.Params{
		PageSize:  2,
		PageToken: result.NextPageToken,
		Category:  category,
	})
	require.NoError(t, err)
	assert.Len(t, result2.Products, 2)
	assert.NotEmpty(t, result2.NextPageToken)
}

func TestOutboxEventCreation(t *testing.T) {
	ctx := context.Background()

	productID := createTestProduct(t, ctx, "Event Test", "electronics")
	now := time.Now().UTC()

	// Apply discount
	require.NoError(t, applyDiscountUC.Execute(ctx, apply_discount.ApplyRequest{
		ProductID:  productID,
		Percentage: big.NewRat(15, 1),
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(24 * time.Hour),
	}))

	// Remove discount
	require.NoError(t, removeDiscountUC.Execute(ctx, apply_discount.RemoveRequest{ProductID: productID}))

	events := getOutboxEvents(t, ctx, productID)
	types := make([]string, len(events))
	for i, e := range events {
		types[i] = e.eventType
	}

	assert.Contains(t, types, "product.created")
	assert.Contains(t, types, "discount.applied")
	assert.Contains(t, types, "discount.removed")

	for _, e := range events {
		assert.Equal(t, "pending", e.status)
		assert.Equal(t, productID, e.aggregateID)
		assert.NotEmpty(t, e.payload)
	}
}

// --- setup helpers ---

func provisionEmulator(ctx context.Context, databaseID string) error {
	instAdmin, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("instance admin: %w", err)
	}
	defer instAdmin.Close()

	instPath := fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID)

	// Attempt to create (ignore AlreadyExists).
	op, err := instAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + projectID,
		InstanceId: instanceID,
		Instance: &instancepb.Instance{
			Config:      fmt.Sprintf("projects/%s/instanceConfigs/emulator-config", projectID),
			DisplayName: "Test",
			NodeCount:   1,
		},
	})
	if err == nil {
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("wait instance: %w", err)
		}
	}

	dbAdmin, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("database admin: %w", err)
	}
	defer dbAdmin.Close()

	ddl, err := os.ReadFile("../../migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	statements := splitDDL(string(ddl))

	dbOp, err := dbAdmin.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          instPath,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
		ExtraStatements: statements,
	})
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	if _, err := dbOp.Wait(ctx); err != nil {
		return fmt.Errorf("wait database: %w", err)
	}

	return nil
}

func splitDDL(raw string) []string {
	parts := strings.Split(raw, ";")
	var out []string
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func wireUsecases(client *spanner.Client) {
	testClock = clock.RealClock{}
	cm := committer.NewCommitter(client)
	productRepo := repo.NewProductRepo(client)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewProductReadModel(client)

	createProductUC = create_product.NewInteractor(productRepo, outboxRepo, cm, testClock)
	updateProductUC = update_product.NewInteractor(productRepo, outboxRepo, cm, testClock)
	applyDiscountUC = apply_discount.NewApplyInteractor(productRepo, outboxRepo, cm, testClock)
	removeDiscountUC = apply_discount.NewRemoveInteractor(productRepo, outboxRepo, cm, testClock)
	activateUC = activate_product.NewActivateInteractor(productRepo, outboxRepo, cm, testClock)
	deactivateUC = activate_product.NewDeactivateInteractor(productRepo, outboxRepo, cm, testClock)
	getProductQuery = get_product.NewHandler(readModel, testClock)
	listProductsQuery = list_products.NewHandler(readModel, testClock)
}

func createTestProduct(t *testing.T, ctx context.Context, name, category string) string {
	t.Helper()
	id, err := createProductUC.Execute(ctx, create_product.Request{
		Name:        name,
		Description: "test product",
		Category:    category,
		BasePrice:   big.NewRat(4999, 100), // $49.99
	})
	require.NoError(t, err)
	return id
}

func createTestProductWithCategory(t *testing.T, ctx context.Context, name, category string) string {
	t.Helper()
	return createTestProduct(t, ctx, name, category)
}

type outboxRow struct {
	eventType   string
	aggregateID string
	status      string
	payload     string
}

func getOutboxEvents(t *testing.T, ctx context.Context, aggregateID string) []outboxRow {
	t.Helper()
	stmt := spanner.Statement{
		SQL: `SELECT event_type, aggregate_id, status, payload
			  FROM outbox_events WHERE aggregate_id = @id ORDER BY created_at`,
		Params: map[string]interface{}{"id": aggregateID},
	}

	iter := spannerClient.Single().Query(ctx, stmt)
	defer iter.Stop()

	var rows []outboxRow
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		require.NoError(t, err)

		var r outboxRow
		var payload json.RawMessage
		require.NoError(t, row.Columns(&r.eventType, &r.aggregateID, &r.status, &payload))
		r.payload = string(payload)
		rows = append(rows, r)
	}
	return rows
}
