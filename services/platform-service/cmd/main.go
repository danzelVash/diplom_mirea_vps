package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"platform-service/config"
	"platform-service/internal/adapters"
	devicegrpc "platform-service/internal/device/app/device/v1"
	devicehttp "platform-service/internal/device/httpapi"
	deviceservice "platform-service/internal/device/service"
	devicestore "platform-service/internal/device/store"
	edgebridgegrpc "platform-service/internal/edgebridge/app/edge_bridge/v1"
	edgebridgehttp "platform-service/internal/edgebridge/httpapi"
	edgebridgeservice "platform-service/internal/edgebridge/service"
	edgebridgestore "platform-service/internal/edgebridge/store"
	gatewaygrpc "platform-service/internal/gateway/app/gateway/v1"
	scenariogrpc "platform-service/internal/scenario/app/scenario/v1"
	scenariohttp "platform-service/internal/scenario/httpapi"
	scenarioservice "platform-service/internal/scenario/service"
	scenariostore "platform-service/internal/scenario/store"
	voiceservice "platform-service/internal/voice"
	voicegrpc "platform-service/internal/voice/app/voice/v1"
	gatewayv1 "platform-service/pkg/pb/api_gateway/v1"
	devicev1 "platform-service/pkg/pb/device/v1"
	edgebridgev1 "platform-service/pkg/pb/edge_bridge/v1"
	scenariov1 "platform-service/pkg/pb/scenario/v1"
	voicev1 "platform-service/pkg/pb/voice/v1"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := ""
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.LoadDefault()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	deviceStore := devicestore.New(pool)
	scenarioStore := scenariostore.New(pool)
	edgeBridgeStore := edgebridgestore.New(pool)

	must(deviceStore.Migrate(ctx))
	must(scenarioStore.Migrate(ctx))
	must(edgeBridgeStore.Migrate(ctx))

	deviceSvc := deviceservice.New(deviceStore)
	deviceGRPC := devicegrpc.New(deviceSvc)
	deviceClient := adapters.DeviceClient{Server: deviceGRPC}

	scenarioSvc := scenarioservice.New(scenarioStore, deviceClient)
	scenarioGRPC := scenariogrpc.New(scenarioSvc)
	scenarioClient := adapters.ScenarioClient{Server: scenarioGRPC}

	recognitionClient, recognitionConn, err := adapters.NewRecognitionClient(cfg.VoiceRecognitionAddr)
	if err != nil {
		log.Fatalf("voice recognition client: %v", err)
	}
	defer recognitionConn.Close()

	voiceSvc := voiceservice.New(scenarioClient, recognitionClient)
	voiceGRPC := voicegrpc.New(voiceSvc)
	voiceClient := adapters.VoiceClient{Server: voiceGRPC}

	edgeBridgeSvc := edgebridgeservice.New(edgeBridgeStore, deviceClient, scenarioClient, voiceClient, cfg.CommandRetryAfter)
	edgeBridgeGRPC := edgebridgegrpc.New(edgeBridgeSvc)
	gatewayGRPC := gatewaygrpc.New(deviceClient, scenarioClient, voiceClient)

	grpcServer := grpc.NewServer()
	devicev1.RegisterDeviceServiceServer(grpcServer, deviceGRPC)
	scenariov1.RegisterScenarioServiceServer(grpcServer, scenarioGRPC)
	edgebridgev1.RegisterEdgeBridgeServiceServer(grpcServer, edgeBridgeGRPC)
	gatewayv1.RegisterGatewayServiceServer(grpcServer, gatewayGRPC)
	voicev1.RegisterVoiceServiceServer(grpcServer, voiceGRPC)

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr())
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           CORSMiddleware(buildHTTPHandler(deviceSvc, scenarioSvc, edgeBridgeSvc)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 2)
	go func() {
		log.Printf("platform-service grpc listening on %s", cfg.GRPCAddr())
		errCh <- grpcServer.Serve(grpcListener)
	}()
	go func() {
		log.Printf("platform-service http listening on %s", cfg.HTTPAddr())
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
}

func buildHTTPHandler(deviceSvc *deviceservice.Service, scenarioSvc *scenarioservice.Service, edgeSvc *edgebridgeservice.Service) http.Handler {
	deviceHandler := devicehttp.New(deviceSvc)
	scenarioHandler := scenariohttp.New(scenarioSvc)
	edgeHandler := edgebridgehttp.New(edgeSvc)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/health":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok","service":"platform-service"}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/edges"):
			edgeHandler.ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/v1/scenarios"):
			scenarioHandler.ServeHTTP(w, r)
		case strings.HasPrefix(r.URL.Path, "/api/v1/devices"),
			strings.HasPrefix(r.URL.Path, "/api/v1/rooms"),
			strings.HasPrefix(r.URL.Path, "/api/v1/inventory"):
			deviceHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
