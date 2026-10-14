package seeders

import (
	"log"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Options struct {
	Force bool
}

type SeedRun struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func SeedData(db *gorm.DB, opts Options) {
	log.Println("Bắt đầu seeding dữ liệu...")
	if err := db.AutoMigrate(&SeedRun{}); err != nil {
		log.Printf("Lỗi khi migrate bảng seed_runs: %v", err)
		return
	}

	if opts.Force {
		log.Println("SEED_FORCE=true: Xoá toàn bộ dữ liệu cũ...")
		truncateAll(db)
		// Reset seed run marker
		db.Where("name = ?", "default").Delete(&SeedRun{})
	}

	if !opts.Force && seedAlreadyRun(db) {
		log.Println("Seed đã chạy trước đó, bỏ qua.")
		return
	}

	seedUniversities(db)
	seedCategories(db)
	seedLocations(db)
	seedShippingServices(db)
	seedAdminUser(db)
	markSeedRun(db)
	log.Println("Seeding hoàn tất (admin + cơ bản)!")
}

func truncateAll(db *gorm.DB) {
	tables := []string{
		"bids", "auction_configs", "appointments",
		"notifications", "offers", "item_images", "item_categories",
		"items", "shipping_services", "locations", "categories",
		"users", "universities", "seed_runs",
	}
	for _, t := range tables {
		db.Exec("DELETE FROM " + t) // ignore errors (table may not exist yet)
	}
	log.Println("Đã xoá tất cả dữ liệu.")
}

func seedAlreadyRun(db *gorm.DB) bool {
	var run SeedRun
	db.Where("name = ?", "default").Limit(1).Find(&run)
	return run.ID != 0
}

func markSeedRun(db *gorm.DB) {
	_ = db.Where("name = ?", "default").FirstOrCreate(&SeedRun{Name: "default"}).Error
}

func seedAdminUser(db *gorm.DB) {
	const (
		adminEmail    = "admin@ecommerce.local"
		adminPassword = "admin123"
		adminName     = "Admin"
	)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Lỗi khi hash mật khẩu admin: %v", err)
		return
	}

	admin := models.User{Email: adminEmail}
	db.Unscoped().Where("email = ?", adminEmail).Limit(1).Find(&admin)
	if admin.ID == 0 {
		admin = models.User{
			Name:       adminName,
			Email:      adminEmail,
			Password:   string(hashedPassword),
			Role:       "admin",
			IsVerified: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Lỗi khi seed admin: %v", err)
			return
		}
	} else if admin.DeletedAt.Valid {
		admin.Name = adminName
		admin.Role = "admin"
		admin.IsVerified = true
		if admin.Password == "" {
			admin.Password = string(hashedPassword)
		}
		admin.DeletedAt.Valid = false
		if err := db.Unscoped().Save(&admin).Error; err != nil {
			log.Printf("Lỗi khi restore admin: %v", err)
			return
		}
	}

	log.Println("- Seed Admin OK (admin@ecommerce.local / admin123)")
}

func seedUsersAndInventory(db *gorm.DB) {
	const defaultPassword = "user123"

	var hust models.University
	_ = db.Where("short_name = ?", "HUST").First(&hust).Error

	users := []models.User{
		{Name: "Minh", Email: "minh@ecommerce.local", Role: "user", IsVerified: true, UniversityID: &hust.ID},
		{Name: "Linh", Email: "linh@ecommerce.local", Role: "user", IsVerified: true, UniversityID: &hust.ID},
	}

	for i := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Lỗi khi hash mật khẩu user %s: %v", users[i].Email, err)
			continue
		}
		users[i].Password = string(hashedPassword)

		existing := models.User{}
		db.Unscoped().Where("email = ?", users[i].Email).Limit(1).Find(&existing)
		if existing.ID == 0 {
			if err := db.Create(&users[i]).Error; err != nil {
				log.Printf("Lỗi khi seed user %s: %v", users[i].Email, err)
				continue
			}
			existing = users[i]
		} else if existing.DeletedAt.Valid {
			existing.Name = users[i].Name
			existing.Role = "user"
			existing.IsVerified = true
			existing.UniversityID = users[i].UniversityID
			if existing.Password == "" {
				existing.Password = users[i].Password
			}
			existing.DeletedAt.Valid = false
			if err := db.Unscoped().Save(&existing).Error; err != nil {
				log.Printf("Lỗi khi restore user %s: %v", users[i].Email, err)
				continue
			}
		}

		seedInventoryForUser(db, &existing)
	}

	log.Println("- Seed Users & Inventory OK")
}

func seedInventoryForUser(db *gorm.DB, user *models.User) {
	itemImageURL := "https://picsum.photos/360"

	var catDienTu, catSach, catGiaDung models.Category
	_ = db.Where("name = ?", "Điện tử").First(&catDienTu).Error
	_ = db.Where("name = ?", "Sách").First(&catSach).Error
	_ = db.Where("name = ?", "Gia dụng").First(&catGiaDung).Error

	items := []models.Item{
		{
			Name:       "Tai nghe Bluetooth",
			UserID:     user.ID,
			Categories: []*models.Category{&catDienTu},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Sách nhập môn lập trình",
			UserID:     user.ID,
			Categories: []*models.Category{&catSach},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Máy sấy tóc",
			UserID:     user.ID,
			Categories: []*models.Category{&catGiaDung},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
	}

	for i := range items {
		existing := models.Item{}
		db.Where("name = ? AND user_id = ?", items[i].Name, user.ID).Limit(1).Find(&existing)
		if existing.ID == 0 {
			if err := db.Create(&items[i]).Error; err != nil {
				log.Printf("Lỗi khi seed item %s: %v", items[i].Name, err)
			}
		}
	}
}

func seedUniversities(db *gorm.DB) {
	universities := []models.University{
		{Name: "Đại học Bách Khoa Hà Nội", ShortName: "HUST"},
		{Name: "Đại học Kinh tế Quốc dân", ShortName: "NEU"},
		{Name: "Đại học Vinh", ShortName: "VinhUni"},
	}

	for _, uni := range universities {
		err := db.Where("short_name = ?", uni.ShortName).FirstOrCreate(&uni).Error
		if err != nil {
			log.Printf("Lỗi khi seed trường %s: %v", uni.ShortName, err)
		}
	}
	log.Println("- Seed Universities OK")
}

func seedCategories(db *gorm.DB) {
	categories := []models.Category{
		{Name: "Điện tử"},
		{Name: "Gia dụng"},
		{Name: "Sách"},
		{Name: "Nội thất"},
		{Name: "Phương tiện"},
		{Name: "Sưu tầm"},
	}

	for _, c := range categories {
		err := db.Where("name = ?", c.Name).FirstOrCreate(&c).Error
		if err != nil {
			log.Printf("Lỗi khi seed trường %s: %v", c.Name, err)
		}
	}

	log.Println("- Seed Categories OK")
}

func seedItems(db *gorm.DB) {
	itemImageURL := "https://picsum.photos/360"

	var owner models.User
	ownerErr := db.Where("email = ?", "admin@ecommerce.local").First(&owner).Error
	if ownerErr != nil {
		ownerErr = db.First(&owner).Error
	}
	if ownerErr != nil {
		log.Printf("Lỗi khi tìm user cho seed items: %v", ownerErr)
		return
	}

	var catDienTu, catSach, catPhuongTien, catGiaDung models.Category
	db.Where("name = ?", "Điện tử").First(&catDienTu)
	db.Where("name = ?", "Sách").First(&catSach)
	db.Where("name = ?", "Phương tiện").First(&catPhuongTien)
	db.Where("name = ?", "Gia dụng").First(&catGiaDung)

	items := []models.Item{
		{
			Name:       "Máy tính cầm tay",
			UserID:     owner.ID,
			Categories: []*models.Category{&catDienTu},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Giáo trình Marx-Lenin",
			UserID:     owner.ID,
			Categories: []*models.Category{&catSach},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Giáo trình đại số tuyến tính",
			UserID:     owner.ID,
			Categories: []*models.Category{&catSach},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Porche 911",
			UserID:     owner.ID,
			Categories: []*models.Category{&catPhuongTien},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
		{
			Name:       "Nồi cơm điện",
			UserID:     owner.ID,
			Categories: []*models.Category{&catGiaDung},
			ItemImage:  []*models.ItemImage{{URL: itemImageURL}},
		},
	}

	for _, item := range items {
		err := db.Where("name = ? AND user_id = ?", item.Name, item.UserID).FirstOrCreate(&item).Error
		if err != nil {
			log.Printf("Lỗi khi seed item %s: %v", item.Name, err)
		}
	}

	log.Println("- Seed Items OK")
}

func seedAuctions(db *gorm.DB) {
	itemImageURL := "https://picsum.photos/480"

	var seller models.User
	if err := db.Where("email = ?", "admin@ecommerce.local").First(&seller).Error; err != nil {
		if err := db.First(&seller).Error; err != nil {
			log.Printf("Lỗi khi tìm seller cho seed auctions: %v", err)
			return
		}
	}

	var bidder1, bidder2 models.User
	_ = db.Where("email = ?", "minh@ecommerce.local").First(&bidder1).Error
	_ = db.Where("email = ?", "linh@ecommerce.local").First(&bidder2).Error

	var catDienTu, catSach models.Category
	_ = db.Where("name = ?", "Điện tử").First(&catDienTu).Error
	_ = db.Where("name = ?", "Sách").First(&catSach).Error

	type auctionSeed struct {
		item    models.Item
		base    uint
		step    uint
		deposit uint
	}

	auctions := []auctionSeed{
		{
			item: models.Item{
				Name:        "Laptop cu dau gia",
				Description: "Laptop cu, may on dinh, dau gia tai thu vien.",
				Type:        models.TypeAuction,
				Status:      models.StatusAvailable,
				UserID:      seller.ID,
				Categories:  []*models.Category{&catDienTu},
				ItemImage:   []*models.ItemImage{{URL: itemImageURL}},
			},
			base:    1500000,
			step:    50000,
			deposit: 200000,
		},
		{
			item: models.Item{
				Name:        "Combo sach on thi",
				Description: "Bo sach on thi cu, phu hop sinh vien nam cuoi.",
				Type:        models.TypeAuction,
				Status:      models.StatusAvailable,
				UserID:      seller.ID,
				Categories:  []*models.Category{&catSach},
				ItemImage:   []*models.ItemImage{{URL: itemImageURL}},
			},
			base:    250000,
			step:    10000,
			deposit: 50000,
		},
	}

	for _, seed := range auctions {
		existing := models.Item{}
		db.Where("name = ? AND user_id = ?", seed.item.Name, seed.item.UserID).Limit(1).Find(&existing)
		if existing.ID == 0 {
			if err := db.Create(&seed.item).Error; err != nil {
				log.Printf("Lỗi khi seed auction item %s: %v", seed.item.Name, err)
				continue
			}
			existing = seed.item
		} else {
			updates := map[string]any{}
			if existing.Type != models.TypeAuction {
				updates["type"] = models.TypeAuction
			}
			if existing.Status == "" {
				updates["status"] = models.StatusAvailable
			}
			if len(updates) > 0 {
				if err := db.Model(&existing).Updates(updates).Error; err != nil {
					log.Printf("Lỗi khi update auction item %s: %v", existing.Name, err)
				}
			}
		}

		var imageCount int64
		_ = db.Model(&models.ItemImage{}).Where("item_id = ?", existing.ID).Count(&imageCount).Error
		if imageCount == 0 {
			_ = db.Create(&models.ItemImage{ItemID: existing.ID, URL: itemImageURL}).Error
		}

		var config models.AuctionConfig
		db.Where("item_id = ?", existing.ID).First(&config)
		if config.ID == 0 {
			config = models.AuctionConfig{
				ItemID:        existing.ID,
				BasePrice:     seed.base,
				StepPrice:     seed.step,
				DepositAmount: seed.deposit,
				Deadline:      time.Now().Add(48 * time.Hour),
				Status:        models.AuctionStatusOpen,
			}
			if err := db.Create(&config).Error; err != nil {
				log.Printf("Lỗi khi seed auction config cho item %s: %v", existing.Name, err)
				continue
			}
		}

		var bidCount int64
		_ = db.Model(&models.Bid{}).Where("auction_config_id = ?", config.ID).Count(&bidCount).Error
		if bidCount == 0 && bidder1.ID != 0 && bidder2.ID != 0 {
			bids := []models.Bid{
				{
					AuctionConfigID: config.ID,
					BidderID:        bidder1.ID,
					Amount:          seed.base + seed.step,
					DepositPaid:     true,
					Status:          models.BidStatusActive,
					Rank:            2,
				},
				{
					AuctionConfigID: config.ID,
					BidderID:        bidder2.ID,
					Amount:          seed.base + seed.step*2,
					DepositPaid:     true,
					Status:          models.BidStatusActive,
					Rank:            1,
				},
			}

			for _, bid := range bids {
				if err := db.Create(&bid).Error; err != nil {
					log.Printf("Lỗi khi seed bid cho item %s: %v", existing.Name, err)
				}
			}
		}
	}

	log.Println("- Seed Auctions OK")
}

func vinhLat(v float64) *float64 { l := 18.673 + v; return &l }
func vinhLng(v float64) *float64 { l := 105.692 + v; return &l }

func seedLocations(db *gorm.DB) {
	locations := []models.Location{
		{Name: "Quảng trường Hồ Chí Minh", Description: "Quảng trường trung tâm TP Vinh, rộng rãi, an ninh tốt.", Address: "Đường Lê Nin, TP Vinh", Latitude: vinhLat(0.005), Longitude: vinhLng(0.002), OpenTime: "06:00", CloseTime: "22:00", IsActive: true},
		{Name: "Công viên Trung tâm", Description: "Công viên lớn, có ghế đá, phù hợp gặp mặt trao đổi.", Address: "Đường Trường Thi, TP Vinh", Latitude: vinhLat(0.003), Longitude: vinhLng(0.001), OpenTime: "05:00", CloseTime: "23:00", IsActive: true},
		{Name: "Đại học Vinh - Cổng chính", Description: "Khu vực an toàn, có bảo vệ, đông sinh viên qua lại.", Address: "182 Lê Duẩn, TP Vinh", Latitude: vinhLat(-0.002), Longitude: vinhLng(0.004), OpenTime: "06:00", CloseTime: "22:00", IsActive: true},
		{Name: "Thư viện tỉnh Nghệ An", Description: "Yên tĩnh, có camera an ninh, phù hợp trao đổi sách vở.", Address: "Đường Nguyễn Thị Minh Khai, TP Vinh", Latitude: vinhLat(0.004), Longitude: vinhLng(-0.001), OpenTime: "07:00", CloseTime: "17:00", IsActive: true},
		{Name: "Siêu thị Big C Vinh", Description: "Khu vực công cộng, đông người, có chỗ để xe rộng.", Address: "Đường Quang Trung, TP Vinh", Latitude: vinhLat(0.001), Longitude: vinhLng(0.006), OpenTime: "08:00", CloseTime: "22:00", IsActive: true},
		{Name: "Quán cà phê Highlands Vinh", Description: "Không gian thoải mái, wifi miễn phí, an toàn.", Address: "Đường Hà Huy Tập, TP Vinh", Latitude: vinhLat(0.006), Longitude: vinhLng(0.003), OpenTime: "07:00", CloseTime: "22:30", IsActive: true},
		{Name: "Bến xe Vinh", Description: "Điểm trung chuyển thuận tiện cho người ở xa, có bảo vệ.", Address: "Đường Lê Viết Thuật, TP Vinh", Latitude: vinhLat(-0.005), Longitude: vinhLng(0.008), OpenTime: "04:00", CloseTime: "23:00", IsActive: true},
		{Name: "Nhà văn hóa Lao động", Description: "Khu vực văn hóa, có sân rộng, an ninh.", Address: "Đường Hồ Tùng Mậu, TP Vinh", Latitude: vinhLat(0.002), Longitude: vinhLng(-0.003), OpenTime: "06:00", CloseTime: "21:00", IsActive: true},
	}

	for _, loc := range locations {
		err := db.Where("name = ?", loc.Name).FirstOrCreate(&loc).Error
		if err != nil {
			log.Printf("Lỗi khi seed location %s: %v", loc.Name, err)
		}
	}
	log.Println("- Seed Locations (Vinh) OK")
}

func seedShippingServices(db *gorm.DB) {
	services := []models.ShippingService{
		{Name: "Giao hàng nhanh", Description: "Giao trong nội thành Vinh, nhận hàng sau 2-4h", BasePrice: 15000, PricePerKm: 3000, MinDistance: 0, EstimatedDays: "0-1", IsActive: true},
		{Name: "Viettel Post", Description: "Chuyển phát toàn quốc, đảm bảo an toàn hàng hóa", BasePrice: 20000, PricePerKm: 2500, MinDistance: 5, EstimatedDays: "1-3", IsActive: true},
		{Name: "Giao hàng tiết kiệm", Description: "Tiết kiệm chi phí, giao trong 2-4 ngày", BasePrice: 12000, PricePerKm: 1800, MinDistance: 10, EstimatedDays: "2-4", IsActive: true},
		{Name: "J&T Express", Description: "Giao hàng nhanh toàn quốc, tracking realtime", BasePrice: 18000, PricePerKm: 3500, MinDistance: 2, EstimatedDays: "0-2", IsActive: true},
		{Name: "Ninja Van", Description: "Dịch vụ giao hàng công nghệ, phù hợp đồ điện tử", BasePrice: 25000, PricePerKm: 4000, MinDistance: 3, EstimatedDays: "1-2", IsActive: true},
	}

	for _, s := range services {
		err := db.Where("name = ?", s.Name).FirstOrCreate(&s).Error
		if err != nil {
			log.Printf("Lỗi khi seed shipping %s: %v", s.Name, err)
		}
	}
	log.Println("- Seed Shipping Services OK")
}
