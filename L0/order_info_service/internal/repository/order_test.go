package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/model"
	"github.com/NikitaKoros/wb_tech/L0/order_info_service/pkg/srvcerrors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host     = "localhost"
	port     = "5431"
	user     = "testuser"
	password = "testpass"
	dbname   = "testdb"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	ConnectTestDB()
	setupTestDBSchema()
	code := m.Run()
	TestDB.Close()
	os.Exit(code)
}

func ConnectTestDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	TestDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	err = TestDB.Ping()
	if err != nil {
		log.Fatalf("failed to verify connection to db: %v", err)
	}
	log.Println("testdb successfully connected")
}

func setupTestDBSchema() {
	schemaPath := "./../../../db/postgres.sql"
	
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("failed to read test_db schema from file: %v", err)
	}
	
	_, err = TestDB.Exec(string(schema))
	if err != nil {
		log.Fatalf("failed to setup test_db schema: %v", err)
	}
}

func clearTables(t *testing.T) {
	_, err := TestDB.Exec("TRUNCATE TABLE items, payments, deliveries, orders RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestCreateOrder_Success(t *testing.T) {
	clearTables(t)
	repo := NewOrderRepository(TestDB)
	ctx := context.Background()

	order := generateTestOrder()

	createdOrder, err := repo.UpsertOrder(ctx, order)

	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	opts := []cmp.Option{
		cmpopts.IgnoreFields(model.Item{}, "ID"),
		cmpopts.IgnoreFields(model.Order{}, "DateCreated"),
	}

	if diff := cmp.Diff(order, createdOrder, opts...); diff != "" {
		t.Errorf("createdOrder mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateOrder_Fail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewOrderRepository(db)
	ctx := context.Background()
	order := generateTestOrder()

	mock.ExpectBegin()
	expectQuery := regexp.QuoteMeta(insertIntoOrdersQuery)
	mock.ExpectQuery(expectQuery).WillReturnError(srvcerrors.ErrDatabase)
	mock.ExpectRollback()

	createdOrder, err := repo.UpsertOrder(ctx, order)

	require.Error(t, err)
	require.Nil(t, createdOrder)
	require.ErrorIs(t, err, srvcerrors.ErrDatabase)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByID_Success(t *testing.T) {
	clearTables(t)
	repo := NewOrderRepository(TestDB)
	ctx := context.Background()

	order := generateTestOrder()
	createdOrder, err := repo.UpsertOrder(ctx, order)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	gotOrder, err := repo.GetOrderByUID(ctx, order.OrderUID)
	require.NoError(t, err)
	require.NotNil(t, gotOrder)
	opts := []cmp.Option{
		cmpopts.IgnoreFields(model.Order{}, "Items"),
	}
	require.True(t, cmp.Equal(createdOrder, gotOrder, opts...))
}

func TestGetOrderByID_Fail(t *testing.T) {
	clearTables(t)
	repo := NewOrderRepository(TestDB)
	ctx := context.Background()
	orderUID := "nonexictent"

	order, err := repo.GetOrderByUID(ctx, orderUID)
	require.Error(t, err)
	require.Nil(t, order)
	require.ErrorIs(t, err, srvcerrors.ErrNotFound)
}

func TestGetAllOrders_Success(t *testing.T) {
	clearTables(t)
	repo := NewOrderRepository(TestDB)
	ctx := context.Background()

	for i := range(3) {
		order := generateTestOrder()
		order.OrderUID = fmt.Sprintf("uid-%d", i)
		order.Delivery.OrderUID = order.OrderUID
		order.Payment.Transaction = order.OrderUID
		for _, item := range order.Items {
			item.OrderUID = order.OrderUID
		}
		_, err := repo.UpsertOrder(ctx, order)
		require.NoError(t, err)
	}

	orders, err := repo.GetAllOrders(ctx, 10)
	require.NoError(t, err)
	require.Len(t, orders, 3)
}

func TestGetAllOrders_Fail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(getAllOrdersQuery)).WillReturnError(srvcerrors.ErrDatabase)
	mock.ExpectRollback()

	orders, err := repo.GetAllOrders(ctx, 10)

	require.Error(t, err)
	require.Nil(t, orders)
	require.ErrorIs(t, err, srvcerrors.ErrDatabase)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetItemsByOrderUID_Success(t *testing.T) {
	clearTables(t)
	repo := NewOrderRepository(TestDB)
	ctx := context.Background()

	order := generateTestOrder()
	order.Items = append(order.Items, []*model.Item{
		{
			OrderUID: order.OrderUID,
			ChrtID:      1003,
			TrackNumber: "track-123",
		},
		{
			OrderUID: order.OrderUID,
			ChrtID:      1004,
			TrackNumber: "track-123",
		},
	}...)
	
	createdOrder, err := repo.UpsertOrder(ctx, order)
	require.NoError(t, err)
	require.NotNil(t, createdOrder)

	items, err := repo.GetItemsByOrderUID(ctx, order.OrderUID, 2, 10)
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, items[0].ID, 3)
	assert.Equal(t, items[1].ID, 4)
}

func TestGetItemsByOrderUID_Fail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewOrderRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(getItemsByOrderUIDQuery)).WillReturnError(srvcerrors.ErrDatabase)
	mock.ExpectRollback()

	items, err := repo.GetItemsByOrderUID(ctx, "any-uid", 0, 10)

	require.Error(t, err)
	require.Nil(t, items)
	require.ErrorIs(t, err, srvcerrors.ErrDatabase)
	require.NoError(t, mock.ExpectationsWereMet())
}

func generateTestOrder() *model.Order {
	return &model.Order{
		OrderUID:          "test-order-uid",
		TrackNumber:       "track-123",
		Entry:             "entry-1",
		Locale:            "en",
		InternalSignature: "signature",
		CustomerID:        "customer-1",
		DeliveryService:   "delivery-xyz",
		Shardkey:          "sh1",
		SmID:              123,
		DateCreated:       time.Now(),
		OofShard:          "os1",

		Delivery: model.Delivery{
			OrderUID: "test-order-uid",
			Name:     "Johny Silverhand",
			Phone:    "+71234567890",
			Zip:      "123456",
			City:     "Moscow",
			Address:  "Lenina 1",
			Region:   "Moscow Region",
			Email:    "test@example.com",
		},

		Payment: model.Payment{
			Transaction:  "test-order-uid",
			RequestID:    "req-001",
			Currency:     "RUB",
			Provider:     "Provider",
			Amount:       5000,
			PaymentDT:    1620000000,
			Bank:         "Tbank",
			DeliveryCost: 300,
			GoodsTotal:   4700,
			CustomFee:    0,
		},

		Items: []*model.Item{
			{
				OrderUID:    "test-order-uid",
				ChrtID:      1001,
				TrackNumber: "track-123",
				Price:       1500,
				RID:         "rid-1",
				Name:        "Item One",
				Sale:        10,
				Size:        "M",
				TotalPrice:  1350,
				NmID:        501,
				Brand:       "BrandA",
				Status:      1,
			},
			{
				OrderUID:    "test-order-uid",
				ChrtID:      1002,
				TrackNumber: "track-123",
				Price:       3200,
				RID:         "rid-2",
				Name:        "Item Two",
				Sale:        0,
				Size:        "L",
				TotalPrice:  3200,
				NmID:        502,
				Brand:       "BrandB",
				Status:      1,
			},
		},
	}
}
