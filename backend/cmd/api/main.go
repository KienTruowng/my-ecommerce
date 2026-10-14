package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/aichat"
	"github.com/AiEmStylix/ecommerce/backend/internal/appointment"
	"github.com/AiEmStylix/ecommerce/backend/internal/auction"
	"github.com/AiEmStylix/ecommerce/backend/internal/auth"
	"github.com/AiEmStylix/ecommerce/backend/internal/cart"
	"github.com/AiEmStylix/ecommerce/backend/internal/categories"
	"github.com/AiEmStylix/ecommerce/backend/internal/config"
	"github.com/AiEmStylix/ecommerce/backend/internal/items"
	"github.com/AiEmStylix/ecommerce/backend/internal/locations"
	"github.com/AiEmStylix/ecommerce/backend/internal/media"
	"github.com/AiEmStylix/ecommerce/backend/internal/middleware"
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"github.com/AiEmStylix/ecommerce/backend/internal/notify"
	"github.com/AiEmStylix/ecommerce/backend/internal/offer"
	"github.com/AiEmStylix/ecommerce/backend/internal/order"
	"github.com/AiEmStylix/ecommerce/backend/internal/shipping"
	"github.com/AiEmStylix/ecommerce/backend/internal/storage"
	"github.com/AiEmStylix/ecommerce/backend/internal/user"
	"github.com/AiEmStylix/ecommerce/backend/internal/wallet"

	"github.com/AiEmStylix/ecommerce/backend/internal/utils"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	_ "github.com/AiEmStylix/ecommerce/backend/docs"
)

func setupAuthModule(userRepo auth.UserRepository, jwtSvc auth.JWTService, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	authService := auth.NewAuthService(userRepo, jwtSvc)
	authHandler := auth.NewAuthHandler(authService)

	requireAuth := authMid.RequireAuth

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/change-password", requireAuth(authHandler.ChangePassword))
}

func setupUserModule(userRepo *user.UserRepository, authMid *auth.AuthMiddleware, mediaService *media.MediaService, mux *http.ServeMux) {
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService, mediaService)

	requireAuth := authMid.RequireAuth
	requireAdmin := authMid.RequireAdmin

	mux.HandleFunc("POST /api/admin/users/{id}/restore", requireAdmin(userHandler.AdminRestoreUserByID))
	mux.HandleFunc("GET /api/admin/users", requireAdmin(userHandler.AdminGetUsers))
	mux.HandleFunc("GET /api/admin/users/{id}", requireAdmin(userHandler.AdminGetUserByID))
	mux.HandleFunc("PATCH /api/admin/users/{id}", requireAdmin(userHandler.AdminUpdateUserByID))
	mux.HandleFunc("DELETE /api/admin/users/{id}", requireAdmin(userHandler.AdminDeleteUserByID))

	mux.HandleFunc("GET /api/users/profile", requireAuth(userHandler.GetUserProfile))
	mux.HandleFunc("PATCH /api/users/profile", requireAuth(userHandler.UpdateUserProfile))
	mux.HandleFunc("POST /api/users/avatar/upload-url", requireAuth(userHandler.GetAvatarUploadURL))
	mux.HandleFunc("POST /api/users/avatar/confirm", requireAuth(userHandler.ConfirmAvatarUpload))
}

func setupCategoryModule(categoryRepo categories.Repository, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	categoryService := categories.NewService(categoryRepo)
	categoryHandler := categories.NewHandler(categoryService)

	mux.HandleFunc("POST /api/categories", categoryHandler.CreateNewCategory)
	mux.HandleFunc("GET /api/categories", categoryHandler.GetAllCategory)
	mux.HandleFunc("GET /api/categories/{id}", categoryHandler.GetCategoryByID)
	mux.HandleFunc("PATCH /api/categories/{id}", categoryHandler.UpdateCategoryByID)
	mux.HandleFunc("DELETE /api/categories/{id}", categoryHandler.DeleteCategoryByID)
}

func setupLocationModule(locationRepo locations.Repository, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	locationService := locations.NewService(locationRepo)
	locationHandler := locations.NewHandler(locationService)

	requireAdmin := authMid.RequireAdmin

	mux.HandleFunc("GET /api/locations", locationHandler.GetLocations)
	mux.HandleFunc("GET /api/locations/{id}", locationHandler.GetLocationByID)
	mux.HandleFunc("POST /api/admin/locations", requireAdmin(locationHandler.CreateLocation))
	mux.HandleFunc("PATCH /api/admin/locations/{id}", requireAdmin(locationHandler.UpdateLocationByID))
	mux.HandleFunc("DELETE /api/admin/locations/{id}", requireAdmin(locationHandler.DeleteLocationByID))
}

func setupItemsModule(itemRepo items.Repository, authMid *auth.AuthMiddleware, mediaService *media.MediaService, mux *http.ServeMux) {
	itemService := items.NewItemsService(itemRepo, mediaService)
	itemHandler := items.NewItemsHandler(itemService)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("GET /api/items/{id}", itemHandler.GetItemByID)
	mux.HandleFunc("GET /api/items", itemHandler.BrowseItems)
	mux.HandleFunc("POST /api/items", requireAuth(itemHandler.CreateItem))
	mux.HandleFunc("PATCH /api/items/{id}", itemHandler.UpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", itemHandler.DeleteItem)
	mux.HandleFunc("GET /api/users/me/items", requireAuth(itemHandler.GetMyItems))
	mux.HandleFunc("POST /api/items/images/presign", itemHandler.RequestUploadImage)
	mux.HandleFunc("POST /api/items/{id}/images", itemHandler.ConfirmAndSaveImage)
	mux.HandleFunc("DELETE /api/item-images/{imageID}", itemHandler.DeleteImage)
}

func setupCartModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	cartRepo := cart.NewRepository(db)
	cartService := cart.NewService(cartRepo, db)
	cartHandler := cart.NewHandler(cartService)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("POST /api/cart", requireAuth(cartHandler.AddToCart))
	mux.HandleFunc("GET /api/cart", requireAuth(cartHandler.GetCart))
	mux.HandleFunc("PATCH /api/cart", requireAuth(cartHandler.UpdateCart))
	mux.HandleFunc("DELETE /api/cart", requireAuth(cartHandler.RemoveFromCart))
}

func setupOrderModule(db *gorm.DB, authMid *auth.AuthMiddleware, notifySvc *notify.Service, mux *http.ServeMux) {
	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService, notifySvc)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("POST /api/checkout", requireAuth(orderHandler.Checkout))
	mux.HandleFunc("GET /api/orders", requireAuth(orderHandler.GetOrders))
	mux.HandleFunc("GET /api/orders/{id}", requireAuth(orderHandler.GetOrder))
	mux.HandleFunc("POST /api/orders/{id}/cancel", requireAuth(orderHandler.CancelOrder))
}

func setupWalletModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	walletRepo := wallet.NewRepository(db)
	walletService := wallet.NewService(walletRepo)
	walletHandler := wallet.NewHandler(walletService)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("GET /api/wallet", requireAuth(walletHandler.GetWalletSummary))
	mux.HandleFunc("GET /api/wallet/transactions", requireAuth(walletHandler.GetWalletTransactions))
	mux.HandleFunc("POST /api/wallet/topup/request", requireAuth(walletHandler.CreateTopupRequest))
	mux.HandleFunc("POST /api/payments/sepay/webhook", walletHandler.SePayWebhook)
}

func setupOfferModule(db *gorm.DB, authMid *auth.AuthMiddleware, notifySvc *notify.Service, mux *http.ServeMux) {
	offerRepo := offer.NewRepository(db)
	offerService := offer.NewService(offerRepo, db)
	offerHandler := offer.NewHandler(offerService, notifySvc)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("POST /api/items/{id}/offers", requireAuth(offerHandler.CreateOffer))
	mux.HandleFunc("GET /api/users/me/offers", requireAuth(offerHandler.GetSentOffers))
	mux.HandleFunc("DELETE /api/offers/{id}", requireAuth(offerHandler.CancelOffer))
	mux.HandleFunc("GET /api/items/{id}/offers", requireAuth(offerHandler.GetItemOffers))
	mux.HandleFunc("PATCH /api/offers/{id}/status", requireAuth(offerHandler.UpdateOfferStatus))
}

func setupAuctionModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	auctionRepo := auction.NewRepository(db)
	auctionService := auction.NewService(auctionRepo, db)
	auctionHandler := auction.NewHandler(auctionService)

	requireAuth := authMid.RequireAuth
	requireAdmin := authMid.RequireAdmin

	mux.HandleFunc("GET /api/items/{id}/auction", auctionHandler.GetAuction)
	mux.HandleFunc("GET /api/items/{id}/bids", requireAuth(auctionHandler.GetBids))
	mux.HandleFunc("POST /api/items/{id}/auction", requireAuth(auctionHandler.ConfigureAuction))
	mux.HandleFunc("POST /api/items/{id}/bids", requireAuth(auctionHandler.PlaceBid))
	mux.HandleFunc("POST /api/admin/auctions/{id}/settle", requireAdmin(auctionHandler.SettleAuction))
}

func setupAppointmentModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	appointmentRepo := appointment.NewRepository(db)
	appointmentService := appointment.NewService(appointmentRepo)
	appointmentHandler := appointment.NewHandler(appointmentService)
	requireAuth := authMid.RequireAuth

	mux.HandleFunc("GET /api/appointments", requireAuth(appointmentHandler.GetMyAppointments))
	mux.HandleFunc("PATCH /api/appointments/{id}", requireAuth(appointmentHandler.UpdateAppointment))
}

func setupShippingModule(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	shippingRepo := shipping.NewRepository(db)
	shippingHandler := shipping.NewHandler(shippingRepo)

	requireAdmin := authMid.RequireAdmin

	mux.HandleFunc("GET /api/shipping", shippingHandler.ListShipping)
	mux.HandleFunc("GET /api/offers/type", shippingHandler.GetOfferType)
	mux.HandleFunc("GET /api/admin/shipping", requireAdmin(shippingHandler.AdminListShipping))
	mux.HandleFunc("POST /api/admin/shipping", requireAdmin(shippingHandler.AdminCreateShipping))
	mux.HandleFunc("PATCH /api/admin/shipping/{id}", requireAdmin(shippingHandler.AdminUpdateShipping))
	mux.HandleFunc("DELETE /api/admin/shipping/{id}", requireAdmin(shippingHandler.AdminDeleteShipping))
}

// @title           Ecommerce API
// @version         1.0
// @description     API cho dự án Ecommerce

// @server http://localhost:3000/api Môi trường Local (Development)
// @server https://api.ecommerce.com/v1 Môi trường Production

// @securityDefinitions.schemas BearerAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Nhập Token theo định dạng: Bearer <token>
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found. Falling back to system environment variables.")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading variables: %v", err)
	}

	gormLog := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.DBSN()), &gorm.Config{Logger: gormLog})
	if err != nil {
		log.Fatal("Cannot connect to Database:", err)
	}
	db.AutoMigrate(&models.University{}, &models.User{}, &models.Category{}, &models.Location{}, &models.Item{}, &models.ItemImage{}, &models.Offer{}, &models.Appointment{}, &models.AuctionConfig{}, &models.Bid{}, &models.ShippingService{}, &models.Notification{}, &models.Cart{}, &models.Order{}, &models.WalletAccount{}, &models.WalletTransaction{}, &models.WalletTopupRequest{}, &models.SepayTransaction{})

	r2Store, err := storage.NewR2Storage(cfg)
	if err != nil {
		log.Fatalf("Không thể khởi tạo R2 Storage: %v", err)
	}

	jwtService, err := auth.NewJWTService(cfg.JWTSecret)
	if err != nil {
		log.Fatalf("Không thể khởi tạo JWT Service: %v", err)
	}

	mediaService := media.NewMediaService(r2Store, cfg)
	authMiddleware := auth.NewAuthMiddleware(jwtService)
	notifyService := notify.NewService(db)
	sharedItemsRepo := items.NewItemsRepository(db)
	sharedUserRepo := user.NewUserRepository(db)
	sharedCategoryRepo := categories.NewRepository(db)
	sharedLocationRepo := locations.NewRepository(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	setupAuthModule(sharedUserRepo, jwtService, authMiddleware, mux)
	setupUserModule(sharedUserRepo, authMiddleware, mediaService, mux)
	setupCategoryModule(sharedCategoryRepo, authMiddleware, mux)
	setupLocationModule(sharedLocationRepo, authMiddleware, mux)
	setupItemsModule(sharedItemsRepo, authMiddleware, mediaService, mux)
	setupCartModule(db, authMiddleware, mux)
	setupOrderModule(db, authMiddleware, notifyService, mux)
	setupWalletModule(db, authMiddleware, mux)
	setupOfferModule(db, authMiddleware, notifyService, mux)
	setupAuctionModule(db, authMiddleware, mux)
	setupAppointmentModule(db, authMiddleware, mux)
	setupShippingModule(db, authMiddleware, mux)
	setupAIChatModule(cfg, authMiddleware, mux)
	setupAdminStatsRoute(db, authMiddleware, mux)

	notifyHandler := notify.NewHandler(notifyService)
	requireAuth := authMiddleware.RequireAuth
	mux.HandleFunc("GET /api/notifications", requireAuth(notifyHandler.GetNotifications))
	mux.HandleFunc("POST /api/notifications/{id}/read", requireAuth(notifyHandler.MarkRead))
	mux.HandleFunc("POST /api/notifications/read-all", requireAuth(notifyHandler.MarkAllRead))
	mux.HandleFunc("GET /api/offers/pending-counts", requireAuth(notifyHandler.PendingOfferCounts))

	log.Println("Server đang chạy tại http://localhost:3000")
	if err := http.ListenAndServe(":3000", middleware.EnableCORS(mux)); err != nil {
		log.Fatalf("Server bị crash: %v", err)
	}
}

func setupAIChatModule(cfg *config.Config, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	var aiHandler *aichat.Handler
	if cfg.DeepSeekAPIKey != "" {
		aiService := aichat.NewService(cfg.DeepSeekAPIKey)
		aiHandler = aichat.NewHandler(aiService)
	} else {
		aiHandler = aichat.NewHandler(nil)
	}

	requireAuth := authMid.RequireAuth
	mux.HandleFunc("POST /api/ai/chat", requireAuth(aiHandler.Chat))
}

func setupAdminStatsRoute(db *gorm.DB, authMid *auth.AuthMiddleware, mux *http.ServeMux) {
	requireAdmin := authMid.RequireAdmin
	mux.HandleFunc("GET /api/admin/stats", requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		stats := getAdminStats(db)
		utils.WriteJSON(w, http.StatusOK, stats)
	}))
}

type AdminStats struct {
	TotalUsers        int64 `json:"total_users"`
	TotalItems        int64 `json:"total_items"`
	ActiveAuctions    int64 `json:"active_auctions"`
	PendingOrders     int64 `json:"pending_orders"`
	TotalAppointments int64 `json:"total_appointments"`
	CompletedTrades   int64 `json:"completed_trades"`
}

func getAdminStats(db *gorm.DB) AdminStats {
	var stats AdminStats
	db.Table("users").Where("deleted_at IS NULL").Count(&stats.TotalUsers)
	db.Table("items").Where("deleted_at IS NULL").Count(&stats.TotalItems)
	db.Table("auction_configs").Where("status = ? AND deleted_at IS NULL", "open").Count(&stats.ActiveAuctions)
	db.Table("orders").Where("order_status = ? AND deleted_at IS NULL", "confirmed").Count(&stats.PendingOrders)
	db.Table("appointments").Where("deleted_at IS NULL").Count(&stats.TotalAppointments)
	db.Table("orders").Where("order_status = ? AND deleted_at IS NULL", "completed").Count(&stats.CompletedTrades)
	return stats
}
