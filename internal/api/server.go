package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
	"go.uber.org/zap"
	"net/http"
)

type DeviceService interface {
	CreateDevice(ctx context.Context, label *string, algorithm domain.Algorithm) (domain.Device, error)
	GetDevices(ctx context.Context) ([]domain.Device, error)
	GetDevice(ctx context.Context, id uuid.UUID) (domain.Device, error)
}

type SignatureService interface {
	SignTransaction(ctx context.Context, deviceID uuid.UUID, data string) (domain.SignedData, error)
	GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]domain.SignedData, error)
}

// Response is the generic API response container.
type Response struct {
	Data interface{} `json:"data"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

type Config struct {
	Host string
	Port int
}

// Server manages HTTP requests and dispatches them to the appropriate services.
type Server struct {
	logger   *zap.SugaredLogger
	config   Config
	validate *validator.Validate

	deviceService    DeviceService
	signatureService SignatureService
}

// NewServer is a factory to instantiate a new Server.
func NewServer(
	logger *zap.SugaredLogger,
	config Config,
	validate *validator.Validate,
	deviceSvc DeviceService,
	signatureSvc SignatureService,
) *Server {
	return &Server{
		logger:           logger,
		config:           config,
		validate:         validate,
		deviceService:    deviceSvc,
		signatureService: signatureSvc,
	}
}

// GetHttpServer returns a new HTTP server instance with all routes registered.
func (s *Server) GetHttpServer() *http.Server {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v0/health", http.HandlerFunc(s.Health))

	mux.Handle("POST /api/v0/devices", http.HandlerFunc(s.CreateDevice))
	mux.Handle("GET /api/v0/devices", http.HandlerFunc(s.GetDevices))
	mux.Handle("GET /api/v0/devices/{id}", http.HandlerFunc(s.GetDevice))

	mux.Handle("POST /api/v0/devices/{id}/signatures", http.HandlerFunc(s.SignTransaction))
	mux.Handle("GET /api/v0/devices/{id}/signatures", http.HandlerFunc(s.GetSignatures))

	// Add middleware
	logMiddleware := LoggingMiddleware(s.logger)
	loggedMux := logMiddleware(mux)

	listenAddr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	return &http.Server{
		Addr:    listenAddr,
		Handler: loggedMux,
	}
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError))) // nolint:errcheck
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes) // nolint:errcheck
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	response := Response{
		Data: data,
	}

	bytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes) // nolint:errcheck
}
