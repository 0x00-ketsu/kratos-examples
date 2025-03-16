package main

import (
	"context"
	"flag"
	"os"
	"time"
	"usermanage/gen/proto/conf"
	"usermanage/internal/pkg/auth"
	"usermanage/internal/pkg/jwt"
	"usermanage/internal/pkg/log/zap"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap/zapcore"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()

	logSensitiveKeys = []string{"password", "passwd", "pwd", "token", "access_token", "refresh_token", "secret"}
)

var (
	defaultJWTKey              = []byte("inPRpgWvweLuK8cv5kIaN5#GIzCllcWa")
	defaultTokenExpireDuration = 2 * time.Hour
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func endUsername() log.Valuer {
	return func(ctx context.Context) any {
		return auth.Username(ctx)
	}
}

func newLogger(c *conf.Log) log.Logger {
	level := zapcore.Level(int(c.Level) - 1)
	zapOpt := zap.Option{
		Level:      level,
		FilePath:   c.FilePath,
		MaxSize:    int(c.MaxSize),
		MaxBackups: int(c.MaxBackups),
		MaxAge:     int(c.MaxAge),
		Compress:   c.Compress,
	}
	logger, err := zap.NewLogger(zapOpt)
	if err != nil {
		panic(err)
	}

	logFilter := log.NewFilter(
		logger,
		log.FilterKey(logSensitiveKeys...),
	)
	return log.With(logFilter,
		"ts", log.Timestamp("2006-01-02T15:04:05.000Z07:00"),
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
		"enduser.name", endUsername(),
	)
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger := newLogger(bc.Log)

	// Initialize JWT with configuration or fallback to defaults
	jwtKey := defaultJWTKey
	tokenDuration := defaultTokenExpireDuration
	if bc.Jwt != nil {
		if len(bc.Jwt.Secret) > 0 {
			jwtKey = []byte(bc.Jwt.Secret)
		}
		if bc.Jwt.ExpireSeconds > 0 {
			tokenDuration = time.Duration(bc.Jwt.ExpireSeconds) * time.Second
		}
	}
	jwt.Initialize(jwtKey, tokenDuration)

	// Initialize data
	data, err := wireData(bc.Data, logger)
	if err != nil {
		panic(err)
	}

	defer data.Cleanup()
	ctx := context.Background()
	if err := data.Migrate(); err != nil {
		panic(err)
	}

	if err := data.InitializeAdminAccount(ctx); err != nil {
		panic(err)
	}

	// Initialize app
	app, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
