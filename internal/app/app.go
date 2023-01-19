package app

import (
	"fmt"
	"log"
	"net"

	"github.com/Asliddin3/image-servis/genproto/image"
	"github.com/Asliddin3/image-servis/internal/controller/service"
	"github.com/Asliddin3/image-servis/pkg/db"
	"github.com/Asliddin3/image-servis/pkg/logger"

	"github.com/Asliddin3/image-servis/config"
	"github.com/Asliddin3/image-servis/internal/controller/storage/imagestore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run(cfg *config.Config) {
	l := logger.New(cfg.LogLevel)

	pgxURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase)

	pg, err := db.New(pgxURL, db.MaxPoolSize(cfg.PGXPoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()
	imageStore := imagestore.NewDiskImageStore("img")

	ImageService := service.NewImageService(l, pg, imageStore)

	lis, err := net.Listen("tcp", ":"+cfg.ImageServicePort)
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - grpcClient.New: %w", err))
	}
	c := grpc.NewServer(
		grpc.MaxConcurrentStreams(100),
	)

	reflection.Register(c)
	image.RegisterImageServiceServer(c, ImageService)

	l.Info("Server is running on" + "port" + ": " + cfg.ImageServicePort)
	if err := c.Serve(lis); err != nil {
		log.Fatal("Error while listening: ", err)
	}
}
