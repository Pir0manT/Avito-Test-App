package main

import (
	"context"
	"errors"
	"log"
	"myapp/config"
	"myapp/models"
	"myapp/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	db, err = gorm.Open(postgres.Open(cfg.PostgresConn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}

	log.Println("Successfully connected to the database")

	// Установка расширения uuid-ossp
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal("Failed to create extension uuid-ossp: ", err)
	}

	// Создание типа organization_type
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type organization_type: ", err)
	}

	// Создание типа service_type
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE service_type AS ENUM ('Construction', 'Delivery', 'Manufacture');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type service_type: ", err)
	}

	// Создание типа status
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE status AS ENUM ('Created', 'Published', 'Closed');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type status: ", err)
	}

	// Создание типа bid_status
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE bid_status AS ENUM ('Created', 'Published', 'Canceled');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type bid_status: ", err)
	}

	// Создание типа author_type
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE author_type AS ENUM ('Organization', 'User');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type author_type: ", err)
	}

	// Создание типа decision_type
	if err := db.Exec(`
        DO $$ BEGIN
            CREATE TYPE decision_type AS ENUM ('Approved', 'Rejected');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;
    `).Error; err != nil {
		log.Fatal("Failed to create type decision_type: ", err)
	}

	// Автоматическая миграция таблиц
	err = db.AutoMigrate(&models.Employee{}, &models.Organization{}, &models.OrganizationResponsible{}, &models.Tender{}, &models.TenderHistory{}, &models.Bid{}, &models.BidHistory{}, &models.Decision{}, &models.Review{})

	// Создание функции для триггера обновления истории тендера
	if err := db.Exec(`
        CREATE OR REPLACE FUNCTION update_tender_history() RETURNS TRIGGER AS $$
        BEGIN
            -- Вставка записи в таблицу истории до изменений
            INSERT INTO tender_histories (id, tender_id, name, description, service_type, status, version, created_at)
            VALUES (uuid_generate_v4(), OLD.id, OLD.name, OLD.description, OLD.service_type, OLD.status, OLD.version, NOW());

            -- Увеличение версии тендера
            NEW.version := OLD.version + 1;

            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
    `).Error; err != nil {
		log.Fatal("Failed to create function update_tender_history: ", err)
	}

	// Создание триггера обновления истории тендера
	if err := db.Exec(`
    DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'tender_update_trigger') THEN
            CREATE TRIGGER tender_update_trigger
            BEFORE UPDATE ON tenders
            FOR EACH ROW
            EXECUTE FUNCTION update_tender_history();
        END IF;
    END $$;
`).Error; err != nil {
		log.Fatal("Failed to create trigger tender_update_trigger: ", err)
	}

	// Создание функции для триггера обновления истории предложений
	if err := db.Exec(`
    CREATE OR REPLACE FUNCTION update_bid_history() RETURNS TRIGGER AS $$
    BEGIN
        -- Вставка записи в таблицу истории до изменений
        INSERT INTO bid_histories (id, bid_id, name, description, status, version, created_at)
        VALUES (uuid_generate_v4(), OLD.id, OLD.name, OLD.description, OLD.status, OLD.version, NOW());

        -- Увеличение версии предложения
        NEW.version := OLD.version + 1;

        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;
`).Error; err != nil {
		log.Fatal("Failed to create function update_bid_history: ", err)
	}

	// Создание триггера обновления истории предложений
	if err := db.Exec(`
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'bid_update_trigger') THEN
        CREATE TRIGGER bid_update_trigger
        BEFORE UPDATE ON bids
        FOR EACH ROW
        EXECUTE FUNCTION update_bid_history();
    END IF;
END $$;
`).Error; err != nil {
		log.Fatal("Failed to create trigger bid_update_trigger: ", err)
	}

}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	r := router.SetupRouter(db)

	srv := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Starting server at %s in %s mode\n", cfg.ServerAddress, gin.Mode())

	// Ожидание сигнала для остановки сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Контекст с таймаутом для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
