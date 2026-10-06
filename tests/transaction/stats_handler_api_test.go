package transaction_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apigw_cache "github.com/MamangRust/monolith-pointofsale-apigateway/cache/gateway_cache"
	apigateway "github.com/MamangRust/monolith-pointofsale-apigateway/handler"
	"github.com/MamangRust/monolith-pointofsale-shared/errors"
	response_api "github.com/MamangRust/monolith-pointofsale-shared/mapper/response/api"

	pbtransaction "github.com/MamangRust/monolith-pointofsale-pb/transaction"
	tests "github.com/MamangRust/monolith-pointofsale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type TransactionStatsApiTestSuite struct {
	tests.BaseTestSuite
	echo       *echo.Echo
	merchantID int
	userID     int
}

func (s *TransactionStatsApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupTransactionService()
	s.SetupOrderService()

	s.echo = echo.New()

	apigateway.NewHandlerTransaction(
		s.echo,
		pbtransaction.NewTransactionQueryServiceClient(s.Conns["transaction"]),
		pbtransaction.NewTransactionCommandServiceClient(s.Conns["transaction"]),
		pbtransaction.NewTransactionStatsServiceClient(s.Conns["transaction"]),
		pbtransaction.NewTransactionStatsByMerchantServiceClient(s.Conns["transaction"]),
		s.Log,
		response_api.NewTransactionResponseMapper(),
		errors.NewApiHandler(s.Obs, s.Log),
		apigw_cache.NewGatewayCache(s.GetCacheStore()),
	)

	ctx := context.Background()
	s.userID = s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)
	catID := s.SeedCategory(ctx)
	prodID := s.SeedProduct(ctx, s.merchantID, catID)
	orderID := s.SeedOrder(ctx, s.userID, s.merchantID, prodID)

	// Seed a successful transaction
	_, err := s.DBPool().Exec(ctx, `
		INSERT INTO transactions (merchant_id, order_id, amount, payment_method, payment_status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		s.merchantID, orderID, 100000, "credit_card", "success", time.Now())
	s.Require().NoError(err)
}

func (s *TransactionStatsApiTestSuite) TestFindMonthStatusSuccess() {
	now := time.Now()
	url := fmt.Sprintf("/api/transaction/monthly-success?year=%d&month=%d", now.Year(), int(now.Month()))
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	s.Equal("success", res["status"])
	s.NotEmpty(res["data"])
}

func (s *TransactionStatsApiTestSuite) TestFindYearStatusSuccess() {
	year := time.Now().Year()
	url := fmt.Sprintf("/api/transaction/yearly-success?year=%d", year)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	s.Equal("success", res["status"])
	s.NotEmpty(res["data"])
}

func (s *TransactionStatsApiTestSuite) TestFindMonthMethodSuccess() {
	now := time.Now()
	url := fmt.Sprintf("/api/transaction/monthly-method-success?year=%d&month=%d", now.Year(), int(now.Month()))
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	s.Equal("success", res["status"])
	s.NotEmpty(res["data"])
}

func (s *TransactionStatsApiTestSuite) TestFindMonthStatusSuccessByMerchant() {
	now := time.Now()
	url := fmt.Sprintf("/api/transaction/merchant/monthly-success?year=%d&month=%d&merchant_id=%d",
		now.Year(), int(now.Month()), s.merchantID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	s.Equal("success", res["status"])
	s.NotEmpty(res["data"])
}

func TestTransactionStatsApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TransactionStatsApiTestSuite))
}
