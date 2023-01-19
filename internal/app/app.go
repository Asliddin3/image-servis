package app

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Asliddin3/image-servis/genproto/image"
	"github.com/Asliddin3/image-servis/internal/controller/service"
	"github.com/Asliddin3/image-servis/pkg/db"
	"github.com/Asliddin3/image-servis/pkg/logger"

	"github.com/Asliddin3/image-servis/config"
	"github.com/Asliddin3/image-servis/internal/controller/storage/imagestore"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ratelimit "github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/juju/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	gatherTime        = 2 * time.Second
	maxUploadDownload = 10
	maxGettingImages  = 100
)

type rateLimiterInterceptor struct {
	TokenBucket *ratelimit.Bucket
}

func (r *rateLimiterInterceptor) Limit() bool {
	fmt.Printf("Token Avail %d \n", r.TokenBucket.Available())

	tokenRes := r.TokenBucket.TakeAvailable(1)
	if tokenRes == 0 {
		fmt.Printf("Reached Rate-Limiting %d \n", r.TokenBucket.Available())
		return true
	}
	return false
}

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
	limiterUploadDownload := &rateLimiterInterceptor{}
	limiterImagesList := &rateLimiterInterceptor{}

	limiterUploadDownload.TokenBucket = ratelimit.NewBucket(gatherTime, int64(maxUploadDownload))
	limiterImagesList.TokenBucket = ratelimit.NewBucket(gatherTime, int64(maxGettingImages))

	c := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ratelimit.UnaryServerInterceptor(limiterImagesList),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ratelimit.StreamServerInterceptor(limiterUploadDownload),
		),
	)

	reflection.Register(c)
	image.RegisterImageServiceServer(c, ImageService)

	l.Info("Server is running on" + "port" + ": " + cfg.ImageServicePort)
	if err := c.Serve(lis); err != nil {
		log.Fatal("Error while listening: ", err)
	}
}
