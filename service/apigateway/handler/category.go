package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gateway_cache "github.com/MamangRust/monolith-pointofsale-apigateway/cache/gateway_cache"
	"github.com/MamangRust/monolith-pointofsale-pkg/logger"
	"github.com/MamangRust/monolith-pointofsale-shared/domain/requests"
	"github.com/MamangRust/monolith-pointofsale-shared/domain/response"
	"github.com/MamangRust/monolith-pointofsale-shared/errors"
	response_api "github.com/MamangRust/monolith-pointofsale-shared/mapper/response/api"

	pbcategory "github.com/MamangRust/monolith-pointofsale-pb/category"
	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/emptypb"
)

type categoryHandleApi struct {
	clientQuery           pbcategory.CategoryQueryServiceClient
	clientCommand         pbcategory.CategoryCommandServiceClient
	clientStats           pbcategory.CategoryStatsServiceClient
	clientStatsById       pbcategory.CategoryStatsByIdServiceClient
	clientStatsByMerchant pbcategory.CategoryStatsByMerchantServiceClient
	logger                logger.LoggerInterface
	mapping               response_api.CategoryResponseMapper
	apiHandler            errors.ApiHandler
	cache                 *gateway_cache.GatewayCache
}

func NewHandlerCategory(
	router *echo.Echo,
	clientQuery pbcategory.CategoryQueryServiceClient,
	clientCommand pbcategory.CategoryCommandServiceClient,
	clientStats pbcategory.CategoryStatsServiceClient,
	clientStatsById pbcategory.CategoryStatsByIdServiceClient,
	clientStatsByMerchant pbcategory.CategoryStatsByMerchantServiceClient,
	logger logger.LoggerInterface,
	mapping response_api.CategoryResponseMapper,
	apiHandler errors.ApiHandler,
	cache *gateway_cache.GatewayCache,
) *categoryHandleApi {
	categoryHandler := &categoryHandleApi{
		clientQuery:           clientQuery,
		clientCommand:         clientCommand,
		clientStats:           clientStats,
		clientStatsById:       clientStatsById,
		clientStatsByMerchant: clientStatsByMerchant,
		logger:                logger,
		mapping:               mapping,
		apiHandler:            apiHandler,
		cache:                 cache,
	}

	routercategory := router.Group("/api/category")

	routercategory.GET("", apiHandler.Handle("get-category-findallcategory", categoryHandler.FindAllCategory))
	routercategory.GET("/:id", apiHandler.Handle("get-category-findbyid", categoryHandler.FindById))
	routercategory.GET("/active", apiHandler.Handle("get-category-findbyactive", categoryHandler.FindByActive))
	routercategory.GET("/trashed", apiHandler.Handle("get-category-findbytrashed", categoryHandler.FindByTrashed))

	routercategory.GET("/monthly-total-pricing", apiHandler.Handle("get-category-findmonthtotalprice", categoryHandler.FindMonthTotalPrice))
	routercategory.GET("/yearly-total-pricing", apiHandler.Handle("get-category-findyeartotalprice", categoryHandler.FindYearTotalPrice))
	routercategory.GET("/merchant/monthly-total-pricing", apiHandler.Handle("get-category-findmonthtotalpricebymerchant", categoryHandler.FindMonthTotalPriceByMerchant))
	routercategory.GET("/merchant/yearly-total-pricing", apiHandler.Handle("get-category-findyeartotalpricebymerchant", categoryHandler.FindYearTotalPriceByMerchant))
	routercategory.GET("/mycategory/monthly-total-pricing", apiHandler.Handle("get-category-findmonthtotalpricebyid", categoryHandler.FindMonthTotalPriceById))
	routercategory.GET("/mycategory/yearly-total-pricing", apiHandler.Handle("get-category-findyeartotalpricebyid", categoryHandler.FindYearTotalPriceById))

	routercategory.GET("/monthly-pricing", apiHandler.Handle("get-category-findmonthprice", categoryHandler.FindMonthPrice))
	routercategory.GET("/yearly-pricing", apiHandler.Handle("get-category-findyearprice", categoryHandler.FindYearPrice))
	routercategory.GET("/merchant/monthly-pricing", apiHandler.Handle("get-category-findmonthpricebymerchant", categoryHandler.FindMonthPriceByMerchant))
	routercategory.GET("/merchant/yearly-pricing", apiHandler.Handle("get-category-findyearpricebymerchant", categoryHandler.FindYearPriceByMerchant))
	routercategory.GET("/mycategory/monthly-pricing", apiHandler.Handle("get-category-findmonthpricebyid", categoryHandler.FindMonthPriceById))
	routercategory.GET("/mycategory/yearly-pricing", apiHandler.Handle("get-category-findyearpricebyid", categoryHandler.FindYearPriceById))

	routercategory.POST("/create", apiHandler.Handle("post-category-create", categoryHandler.Create))
	routercategory.POST("/update/:id", apiHandler.Handle("post-category-update", categoryHandler.Update))

	routercategory.POST("/trashed/:id", apiHandler.Handle("post-category-trashedcategory", categoryHandler.TrashedCategory))
	routercategory.POST("/restore/:id", apiHandler.Handle("post-category-restorecategory", categoryHandler.RestoreCategory))
	routercategory.DELETE("/permanent/:id", apiHandler.Handle("delete-category-deletecategorypermanent", categoryHandler.DeleteCategoryPermanent))

	routercategory.POST("/restore/all", apiHandler.Handle("post-category-restoreallcategory", categoryHandler.RestoreAllCategory))
	routercategory.POST("/permanent/all", apiHandler.Handle("post-category-deleteallcategorypermanent", categoryHandler.DeleteAllCategoryPermanent))

	return categoryHandler
}

// List all category (paginated)
// List all category (paginated)
// List all category (paginated)
// @Summary List all category (paginated)
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindAllCategory(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findallcategory:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategory](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &pbcategory.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindAll(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategory(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category by ID
// Get category by ID
// Get category by ID
// @Summary Get category by ID
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/:id [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	cacheKey := fmt.Sprintf("category:findbyid:id_%d", id)
	if cached, found := gateway_cache.Get[response.ApiResponseCategory](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &pbcategory.FindByIdCategoryRequest{
		Id: int32(id),
	}

	res, err := h.clientQuery.FindById(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List active category
// List active category
// List active category
// @Summary List active category
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/active [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindByActive(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findbyactive:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategoryDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &pbcategory.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientQuery.FindByActive(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategoryDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// List trashed category
// List trashed category
// List trashed category
// @Summary List trashed category
// @Tags Category
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param search query string false "Search keyword"
// @Success 200 {object} response.ApiResponsePaginationCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/trashed [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindByTrashed(c echo.Context) error {
	ctx := c.Request().Context()
	page := parseQueryInt(c, "page", 1)
	pageSize := parseQueryInt(c, "page_size", 10)
	search := c.QueryParam("search")

	cacheKey := fmt.Sprintf("category:findbytrashed:page_%d:size_%d:search_%s", page, pageSize, search)
	if cached, found := gateway_cache.Get[response.ApiResponsePaginationCategoryDeleteAt](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	req := &pbcategory.FindAllCategoryRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Search:   search,
	}

	res, err := h.clientCommand.FindByTrashed(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponsePaginationCategoryDeleteAt(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/monthly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthTotalPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	month, err := parseQueryIntWithValidation(c, "month", 1, 12)
	if err != nil {
		return errors.NewBadRequestError("invalid month")
	}

	cacheKey := fmt.Sprintf("category:findmonthtotalprice:year_%d:month_%d", year, month)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStats.FindMonthlyTotalPrices(ctx, &pbcategory.FindYearMonthTotalPrices{
		Year:  int32(year),
		Month: int32(month),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/yearly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearTotalPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}

	cacheKey := fmt.Sprintf("category:findyeartotalprice:year_%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStats.FindYearlyTotalPrices(ctx, &pbcategory.FindYearTotalPrices{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/mycategory/monthly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthTotalPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	month, err := parseQueryIntWithValidation(c, "month", 1, 12)
	if err != nil {
		return errors.NewBadRequestError("invalid month")
	}
	categoryID, err := parseQueryIntWithValidation(c, "category_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid category_id")
	}

	cacheKey := fmt.Sprintf("category:findmonthtotalpricebyid:year_%d:month_%d:categoryID_%d", year, month, categoryID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsById.FindMonthlyTotalPricesById(ctx, &pbcategory.FindYearMonthTotalPriceById{
		Year:       int32(year),
		Month:      int32(month),
		CategoryId: int32(categoryID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/mycategory/yearly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearTotalPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	categoryID, err := parseQueryIntWithValidation(c, "category_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid category_id")
	}

	cacheKey := fmt.Sprintf("category:findyeartotalpricebyid:year_%d:categoryID_%d", year, categoryID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsById.FindYearlyTotalPricesById(ctx, &pbcategory.FindYearTotalPriceById{
		Year:       int32(year),
		CategoryId: int32(categoryID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/merchant/monthly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthTotalPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	month, err := parseQueryIntWithValidation(c, "month", 1, 12)
	if err != nil {
		return errors.NewBadRequestError("invalid month")
	}
	merchantID, err := parseQueryIntWithValidation(c, "merchant_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid merchant_id")
	}

	cacheKey := fmt.Sprintf("category:findmonthtotalpricebymerchant:year_%d:month_%d:merchantID_%d", year, month, merchantID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsByMerchant.FindMonthlyTotalPricesByMerchant(ctx, &pbcategory.FindYearMonthTotalPriceByMerchant{
		Year:       int32(year),
		Month:      int32(month),
		MerchantId: int32(merchantID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearlyTotalPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/merchant/yearly-total-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearTotalPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	merchantID, err := parseQueryIntWithValidation(c, "merchant_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid merchant_id")
	}

	cacheKey := fmt.Sprintf("category:findyeartotalpricebymerchant:year_%d:merchantID_%d", year, merchantID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsByMerchant.FindYearlyTotalPricesByMerchant(ctx, &pbcategory.FindYearTotalPriceByMerchant{
		Year:       int32(year),
		MerchantId: int32(merchantID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyTotalPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/monthly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}

	cacheKey := fmt.Sprintf("category:findmonthprice:year_%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStats.FindMonthPrice(ctx, &pbcategory.FindYearCategory{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/yearly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearPrice(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}

	cacheKey := fmt.Sprintf("category:findyearprice:year_%d", year)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStats.FindYearPrice(ctx, &pbcategory.FindYearCategory{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/merchant/monthly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	merchantID, err := parseQueryIntWithValidation(c, "merchant_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid merchant_id")
	}

	cacheKey := fmt.Sprintf("category:findmonthpricebymerchant:year_%d:merchantID_%d", year, merchantID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsByMerchant.FindMonthPriceByMerchant(ctx, &pbcategory.FindYearCategoryByMerchant{
		Year:       int32(year),
		MerchantId: int32(merchantID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/merchant/yearly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearPriceByMerchant(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	merchantID, err := parseQueryIntWithValidation(c, "merchant_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid merchant_id")
	}

	cacheKey := fmt.Sprintf("category:findyearpricebymerchant:year_%d:merchantID_%d", year, merchantID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsByMerchant.FindYearPriceByMerchant(ctx, &pbcategory.FindYearCategoryByMerchant{
		Year:       int32(year),
		MerchantId: int32(merchantID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryMonthPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/mycategory/monthly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindMonthPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	categoryID, err := parseQueryIntWithValidation(c, "category_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid categoryID")
	}

	cacheKey := fmt.Sprintf("category:findmonthpricebyid:year_%d:categoryID_%d", year, categoryID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryMonthPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsById.FindMonthPriceById(ctx, &pbcategory.FindYearCategoryById{
		Year:       int32(year),
		CategoryId: int32(categoryID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryMonthlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Get category statistics
// Get category statistics
// Get category statistics
// @Summary Get category statistics
// @Tags Category
// @Accept json
// @Produce json
// @Param year query int false "Year (e.g. 2026)"
// @Param month query int false "Month (1-12)"
// @Success 200 {object} response.ApiResponseCategoryYearPrice
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/mycategory/yearly-pricing [get]
// @Security BearerAuth
func (h *categoryHandleApi) FindYearPriceById(c echo.Context) error {
	ctx := c.Request().Context()
	year, err := parseQueryIntWithValidation(c, "year", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid year")
	}
	categoryID, err := parseQueryIntWithValidation(c, "category_id", 1, 9999)
	if err != nil {
		return errors.NewBadRequestError("invalid categoryID")
	}

	cacheKey := fmt.Sprintf("category:findyearpricebyid:year_%d:categoryID_%d", year, categoryID)
	if cached, found := gateway_cache.Get[response.ApiResponseCategoryYearPrice](ctx, h.cache, cacheKey); found && cached != nil {
		return c.JSON(http.StatusOK, cached)
	}

	res, err := h.clientStatsById.FindYearPriceById(ctx, &pbcategory.FindYearCategoryById{
		Year:       int32(year),
		CategoryId: int32(categoryID),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryYearlyPrice(res)

	gateway_cache.Set(ctx, h.cache, cacheKey, so, 5*time.Minute)
	return c.JSON(http.StatusOK, so)
}

// Create category
// Create category
// Create category
// @Summary Create category
// @Tags Category
// @Accept json
// @Produce json
// @Param request body requests.CreateCategoryRequest true "Request body"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/create [post]
// @Security BearerAuth
func (h *categoryHandleApi) Create(c echo.Context) error {
	ctx := c.Request().Context()
	var body requests.CreateCategoryRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &pbcategory.CreateCategoryRequest{
		Name:        body.Name,
		Description: body.Description,
	}

	res, err := h.clientCommand.Create(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusCreated, so)
}

// Update category
// Update category
// Update category
// @Summary Update category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Param request body requests.UpdateCategoryRequest true "Request body"
// @Success 200 {object} response.ApiResponseCategory
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/update/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) Update(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	var body requests.UpdateCategoryRequest
	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("bind failed")
	}
	if err := body.Validate(); err != nil {
		return errors.NewBadRequestError("validation failed")
	}

	req := &pbcategory.UpdateCategoryRequest{
		CategoryId:  int32(id),
		Name:        body.Name,
		Description: body.Description,
	}

	res, err := h.clientCommand.Update(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategory(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Trash category
// Trash category
// Trash category
// @Summary Trash category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/trashed/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) TrashedCategory(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &pbcategory.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.TrashedCategory(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Restore category
// Restore category
// Restore category
// @Summary Restore category
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDeleteAt
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/restore/:id [post]
// @Security BearerAuth
func (h *categoryHandleApi) RestoreCategory(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &pbcategory.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.RestoreCategory(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDeleteAt(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Delete category permanently
// Delete category permanently
// Delete category permanently
// @Summary Delete category permanently
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "id"
// @Success 200 {object} response.ApiResponseCategoryDelete
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/permanent/:id [delete]
// @Security BearerAuth
func (h *categoryHandleApi) DeleteCategoryPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return errors.NewBadRequestError("invalid category ID")
	}

	req := &pbcategory.FindByIdCategoryRequest{Id: int32(id)}
	res, err := h.clientCommand.DeleteCategoryPermanent(ctx, req)
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryDelete(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Restore all trashed category
// Restore all trashed category
// Restore all trashed category
// @Summary Restore all trashed category
// @Tags Category
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/restore/all [post]
// @Security BearerAuth
func (h *categoryHandleApi) RestoreAllCategory(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.RestoreAllCategory(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}

// Delete all trashed category permanently
// Delete all trashed category permanently
// Delete all trashed category permanently
// @Summary Delete all trashed category permanently
// @Tags Category
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseCategoryAll
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/category/permanent/all [post]
// @Security BearerAuth
func (h *categoryHandleApi) DeleteAllCategoryPermanent(c echo.Context) error {
	ctx := c.Request().Context()
	res, err := h.clientCommand.DeleteAllCategoryPermanent(ctx, &emptypb.Empty{})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapping.ToApiResponseCategoryAll(res)

	gateway_cache.InvalidatePattern(ctx, h.cache, "category:*")
	return c.JSON(http.StatusOK, so)
}
