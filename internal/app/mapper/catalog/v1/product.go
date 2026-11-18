package mcatv1

import (
	"math"

	"github.com/iFreezy/catalog-service/internal/app/entity"
	catalogv1 "github.com/iFreezy/catalog-service/internal/pkg/grpc/gen/catalog/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProductToProto(p entity.Product) *catalogv1.Product {
	description := ""
	if p.Description != nil {
		description = *p.Description
	}
	return &catalogv1.Product{
		Guid:         p.GUID.String(),
		Name:         p.Name,
		Description:  description,
		Price:        int64(math.Round(p.Price * 100)),
		CategoryGuid: p.CategoryGUID.String(),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}
}
