package services

import (
	"github.com/Nap20192/hacknu/gen/sqlc"
)

type agregatorService struct {
	queries sqlc.Querier
}

func NewAgregatorService(queries sqlc.Querier) *agregatorService {
	return &agregatorService{queries: queries}
}

type agregatorWorker struct {

}
