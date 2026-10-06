package handler

import (
	"fmt"
	pb "github.com/MamangRust/monolith-pointofsale-pb"
	"strconv"

	auth_cache "github.com/MamangRust/monolith-pointofsale-apigateway/cache/auth"
	gateway_cache "github.com/MamangRust/monolith-pointofsale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/monolith-pointofsale-pkg/auth"
	"github.com/MamangRust/monolith-pointofsale-pkg/logger"
	"github.com/MamangRust/monolith-pointofsale-pkg/upload_image"
	"github.com/MamangRust/monolith-pointofsale-shared/errors"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"

	pbcashier "github.com/MamangRust/monolith-pointofsale-pb/cashier"
	pbcategory "github.com/MamangRust/monolith-pointofsale-pb/category"
	pbmerchant "github.com/MamangRust/monolith-pointofsale-pb/merchant"
	pbmerchant_document "github.com/MamangRust/monolith-pointofsale-pb/merchant_document"
	pborder "github.com/MamangRust/monolith-pointofsale-pb/order"
	pbproduct "github.com/MamangRust/monolith-pointofsale-pb/product"
	pbrole "github.com/MamangRust/monolith-pointofsale-pb/role"
	pbtransaction "github.com/MamangRust/monolith-pointofsale-pb/transaction"
	pbuser "github.com/MamangRust/monolith-pointofsale-pb/user"
	response_api "github.com/MamangRust/monolith-pointofsale-shared/mapper/response/api"
)

type ServiceConnections struct {
	Auth        *grpc.ClientConn
	Role        *grpc.ClientConn
	User        *grpc.ClientConn
	Cashier     *grpc.ClientConn
	Category    *grpc.ClientConn
	Merchant    *grpc.ClientConn
	OrderItem   *grpc.ClientConn
	Order       *grpc.ClientConn
	Product     *grpc.ClientConn
	Transaction *grpc.ClientConn
}

type Deps struct {
	Token              auth.TokenManager
	E                  *echo.Echo
	Logger             logger.LoggerInterface
	Mapping            *response_api.ResponseApiMapper
	ImageUpload        upload_image.ImageUploads
	ServiceConnections *ServiceConnections
	ApiHandler         errors.ApiHandler
	AuthCache          auth_cache.AuthMencache
	GatewayCache       *gateway_cache.GatewayCache
}

func NewHandler(deps *Deps) {
	clientAuth := pb.NewAuthServiceClient(deps.ServiceConnections.Auth)
	clientRoleQuery := pbrole.NewRoleQueryServiceClient(deps.ServiceConnections.Role)
	clientRoleCommand := pbrole.NewRoleCommandServiceClient(deps.ServiceConnections.Role)
	clientUserQuery := pbuser.NewUserQueryServiceClient(deps.ServiceConnections.User)
	clientUserCommand := pbuser.NewUserCommandServiceClient(deps.ServiceConnections.User)
	clientCategoryQuery := pbcategory.NewCategoryQueryServiceClient(deps.ServiceConnections.Category)
	clientCategoryCommand := pbcategory.NewCategoryCommandServiceClient(deps.ServiceConnections.Category)
	clientCategoryStats := pbcategory.NewCategoryStatsServiceClient(deps.ServiceConnections.Category)
	clientCategoryStatsById := pbcategory.NewCategoryStatsByIdServiceClient(deps.ServiceConnections.Category)
	clientCategoryStatsByMerchant := pbcategory.NewCategoryStatsByMerchantServiceClient(deps.ServiceConnections.Category)
	clientCashierQuery := pbcashier.NewCashierQueryServiceClient(deps.ServiceConnections.Cashier)
	clientCashierCommand := pbcashier.NewCashierCommandServiceClient(deps.ServiceConnections.Cashier)
	clientCashierStats := pbcashier.NewCashierStatsServiceClient(deps.ServiceConnections.Cashier)
	clientCashierStatsById := pbcashier.NewCashierStatsByIdServiceClient(deps.ServiceConnections.Cashier)
	clientCashierStatsByMerchant := pbcashier.NewCashierStatsByMerchantServiceClient(deps.ServiceConnections.Cashier)
	clientMerchantQuery := pbmerchant.NewMerchantQueryServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantCommand := pbmerchant.NewMerchantCommandServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDocumentQuery := pbmerchant_document.NewMerchantDocumentQueryServiceClient(deps.ServiceConnections.Merchant)
	clientMerchantDocumentCommand := pbmerchant_document.NewMerchantDocumentCommandServiceClient(deps.ServiceConnections.Merchant)
	clientOrderItemQuery := pb.NewOrderItemQueryServiceClient(deps.ServiceConnections.OrderItem)
	clientOrderQuery := pborder.NewOrderQueryServiceClient(deps.ServiceConnections.Order)
	clientOrderCommand := pborder.NewOrderCommandServiceClient(deps.ServiceConnections.Order)
	clientOrderStats := pborder.NewOrderStatsServiceClient(deps.ServiceConnections.Order)
	clientOrderStatsByMerchant := pborder.NewOrderStatsByMerchantServiceClient(deps.ServiceConnections.Order)
	clientProductQuery := pbproduct.NewProductQueryServiceClient(deps.ServiceConnections.Product)
	clientProductCommand := pbproduct.NewProductCommandServiceClient(deps.ServiceConnections.Product)
	clientTransactionQuery := pbtransaction.NewTransactionQueryServiceClient(deps.ServiceConnections.Transaction)
	clientTransactionCommand := pbtransaction.NewTransactionCommandServiceClient(deps.ServiceConnections.Transaction)
	clientTransactionStats := pbtransaction.NewTransactionStatsServiceClient(deps.ServiceConnections.Transaction)
	clientTransactionStatsByMerchant := pbtransaction.NewTransactionStatsByMerchantServiceClient(deps.ServiceConnections.Transaction)

	NewHandlerAuth(deps.E, clientAuth, deps.Logger, deps.Mapping.AuthResponseMapper, deps.ApiHandler, deps.AuthCache)
	NewHandlerRole(deps.E, clientRoleQuery, clientRoleCommand, deps.Logger, deps.Mapping.RoleResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerUser(deps.E, clientUserQuery, clientUserCommand, deps.Logger, deps.Mapping.UserResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCategory(deps.E, clientCategoryQuery, clientCategoryCommand, clientCategoryStats, clientCategoryStatsById, clientCategoryStatsByMerchant, deps.Logger, deps.Mapping.CategoryResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerCashier(deps.E, clientCashierQuery, clientCashierCommand, clientCashierStats, clientCashierStatsById, clientCashierStatsByMerchant, deps.Logger, deps.Mapping.CashierResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchant(deps.E, clientMerchantQuery, clientMerchantCommand, deps.Logger, deps.Mapping.MerchantResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerMerchantDocument(deps.E, clientMerchantDocumentQuery, clientMerchantDocumentCommand, deps.Logger, deps.Mapping.MerchantDocumentProMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrderItem(deps.E, clientOrderItemQuery, deps.Logger, deps.Mapping.OrderItemResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerOrder(deps.E, clientOrderQuery, clientOrderCommand, clientOrderStats, clientOrderStatsByMerchant, deps.Logger, deps.Mapping.OrderResponseMapper, deps.ApiHandler, deps.GatewayCache)
	NewHandlerProduct(deps.E, clientProductQuery, clientProductCommand, deps.Logger, deps.Mapping.ProductResponseMapper, deps.ImageUpload, deps.ApiHandler, deps.GatewayCache)
	NewHandlerTransaction(deps.E, clientTransactionQuery, clientTransactionCommand, clientTransactionStats, clientTransactionStatsByMerchant, deps.Logger, deps.Mapping.TransactionResponseMapper, deps.ApiHandler, deps.GatewayCache)
}

func parseQueryInt(c echo.Context, key string, defaultValue int) int {
	val, err := strconv.Atoi(c.QueryParam(key))
	if err != nil || val <= 0 {
		return defaultValue
	}
	return val
}

func parseQueryIntWithValidation(c echo.Context, key string, min, max int) (int, error) {
	valStr := c.QueryParam(key)
	val, err := strconv.Atoi(valStr)
	if err != nil || val < min || val > max {
		return 0, fmt.Errorf("invalid %s: %s", key, valStr)
	}
	return val, nil
}
