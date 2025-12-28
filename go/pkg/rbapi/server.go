package rbapi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/oklog/run"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	chilogger "github.com/treastech/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
	"rslbot.com/go/pkg/errcode"
)

const (
	MaxMessageSize = 10000 * 1024 * 1024
)

type Server struct {
	grpcServer     *grpc.Server
	grpcListener   net.Listener
	httpListener   net.Listener
	masterListener net.Listener
	cmux           cmux.CMux
	logger         *zap.Logger
	workers        run.Group
	ctx            context.Context
	cancel         func()
}

type ServerOpts struct {
	Logger             *zap.Logger
	Bind               string
	CORSAllowedOrigins string
	RequestTimeout     time.Duration
	ShutdownTimeout    time.Duration
	WithPprof          bool
}

func NewServer(ctx context.Context, svc Service, db *gorm.DB, redisStore *RedisStore, opts ServerOpts) (*Server, error) {
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	if opts.CORSAllowedOrigins == "" {
		opts.CORSAllowedOrigins = "*"
	}
	if opts.RequestTimeout == 0 {
		opts.RequestTimeout = 20 * time.Minute
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 21 * time.Minute
	}

	ctx, cancel := context.WithCancel(ctx)
	s := Server{
		logger: opts.Logger,
		ctx:    ctx,
		cancel: cancel,
	}

	var err error
	s.masterListener, err = net.Listen("tcp", opts.Bind)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	s.cmux = cmux.New(s.masterListener)
	s.grpcListener = s.cmux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	s.httpListener = s.cmux.Match(cmux.HTTP2(), cmux.HTTP1())

	// gRPC server
	s.grpcServer = grpcServer(svc, opts)
	s.workers.Add(func() error {
		return s.grpcServer.Serve(s.grpcListener)
	}, func(error) {
		if err := s.grpcListener.Close(); err != nil {
			opts.Logger.Warn("close listener", zap.Error(err))
		}
	})

	// HTTP server
	httpServer, err := httpServer(ctx, db, redisStore, s.ListenerAddr(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}
	s.workers.Add(func() error {
		return httpServer.Serve(s.httpListener)
	}, func(error) {
		ctx, cancel := context.WithTimeout(ctx, opts.ShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			opts.Logger.Warn("shutdown HTTP server", zap.Error(err))
		}
		if err := s.httpListener.Close(); err != nil {
			opts.Logger.Warn("close listener", zap.Error(err))
		}
	})

	// cmux
	s.workers.Add(func() error {
		return s.cmux.Serve()
	}, func(err error) {
		s.logger.Error("cmux serve error", zap.Error(err))
	})

	// Listen for canceled context
	s.workers.Add(func() error {
		<-s.ctx.Done()
		return nil
	}, func(error) {})

	return &s, nil
}

func (s *Server) Run() error {
	return s.workers.Run()
}

func (s *Server) Close() {
	if s.masterListener != nil {
		s.masterListener.Close()
	}
	s.cancel()
}

func (s *Server) ListenerAddr() string {
	return s.masterListener.Addr().String()
}

func grpcServer(svc Service, opts ServerOpts) *grpc.Server {
	logger := opts.Logger.Named("grpc")
	recoveryOpts := []grpc_recovery.Option{}

	// Create auth function adapter
	authFunc := func(ctx context.Context) (context.Context, error) {
		// Get full method name from gRPC context
		fullMethodName, ok := grpc.Method(ctx)
		if !ok {
			return nil, errcode.ERR_AUTH_MISSING_METADATA
		}
		return svc.(*service).AuthFuncOverride(ctx, fullMethodName)
	}

	// Stream interceptors
	serverStreamOpts := []grpc.StreamServerInterceptor{
		grpc_auth.StreamServerInterceptor(authFunc),
		grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		grpc_zap.StreamServerInterceptor(logger),
	}

	// Unary interceptors
	serverUnaryOpts := []grpc.UnaryServerInterceptor{
		grpc_auth.UnaryServerInterceptor(authFunc),
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		grpc_zap.UnaryServerInterceptor(logger),
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(serverStreamOpts...)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(serverUnaryOpts...)),
		grpc.MaxRecvMsgSize(MaxMessageSize),
	)
	RegisterServiceServer(grpcServer, svc)

	return grpcServer
}

func httpServer(ctx context.Context, db *gorm.DB, redisStore *RedisStore, serverListenerAddr string, opts ServerOpts) (*http.Server, error) {
	logger := opts.Logger.Named("http")

	r := chi.NewRouter()

	// CORS
	cors := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(opts.CORSAllowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r.Use(cors.Handler)
	r.Use(chilogger.Logger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(opts.RequestTimeout))
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// gRPC-Gateway
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated:   false,
					EmitDefaultValues: true,
					UseEnumNumbers:    false,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxMessageSize)),
	}

	err := RegisterServiceHandlerFromEndpoint(ctx, gwmux, serverListenerAddr, dialOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	r.Mount("/", gwmux)
	r.HandleFunc("/license/activate", activateLicense(db, redisStore))
	r.HandleFunc("/license/check", checkLicense(db, redisStore))
	r.HandleFunc("/offsets/update", updateOffsets(db))
	r.HandleFunc("/webhooks/paypal", paypalWebhookHandler(db, logger))
	if opts.WithPprof {
		r.HandleFunc("/debug/pprof/*", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	http.DefaultServeMux = http.NewServeMux() // disables default handlers registered by importing net/http/pprof for security reasons

	return &http.Server{
		Handler: r,
	}, nil
}
