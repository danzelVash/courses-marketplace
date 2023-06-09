package main

import (
	_ "github.com/jackc/pgx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"mom"
	"mom/internal/libraries/logging"
	"mom/internal/libraries/storage"
	"mom/internal/pkg/handler"
	"mom/internal/pkg/repository"
	"mom/internal/pkg/service"
	"os"
)

func main() {
	logger := logging.GetLogger()
	if err := initConfig(); err != nil {
		logger.Fatalf("error reading config file %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logger.Fatal("can`t read db password from env")
	}

	db, err := repository.NewPostgresDb(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		logger.Fatalf("error init db: %s", err.Error())
	}

	strg := storage.NewDataStorage()

	repos := repository.NewRepository(db, strg, logger)
	services := service.NewService(repos, logger)
	handlers := handler.NewHandler(services, logger)

	go services.Authorization.CleanExpiredSessions()

	srv := &courses.Server{}
	if err := srv.Run(viper.GetString("port"), handlers.InitRouters()); err != nil {
		logger.Fatalf("error while trying run server: %s", err.Error())
	}

}

func initConfig() error {
	viper.AddConfigPath("internal/pkg/configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
