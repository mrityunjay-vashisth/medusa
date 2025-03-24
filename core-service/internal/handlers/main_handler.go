package handlers

// import (
// 	"context"
// 	"encoding/json"
// 	"net/http"

// 	"github.com/mrityunjay-vashisth/core-service/internal/db"
// 	"github.com/mrityunjay-vashisth/core-service/internal/registry"
// 	"go.uber.org/zap"
// )

// type HandlerManagerInterface interface {
// 	GetSubHandler(name string) http.HandlerFunc
// }

// // MainHandler manages subhandlers dynamically
// type handlerManager struct {
// 	db       db.DBClientInterface
// 	registry registry.ServiceRegistry
// 	logger   *zap.Logger
// }

// // NewMainHandler initializes subhandlers
// func NewMainHandler(ctx context.Context, db db.DBClientInterface, registry registry.ServiceRegistry) HandlerManagerInterface {
// 	logger, ok := ctx.Value("logger").(*zap.Logger)
// 	if !ok {
// 		logger = zap.L()
// 	}

// 	return &handlerManager{
// 		db:       db,
// 		registry: registry,
// 		logger:   logger,
// 	}
// }

// func respondWithError(w http.ResponseWriter, statusCode int, message string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	json.NewEncoder(w).Encode(map[string]string{"message": message})
// }
