package service

import (
	"context"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/cache"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/LorenzattiGabriel/threat-intel-api/internal/repository"
)

type IndicatorService struct {
	repo  *repository.IndicatorRepository
	cache *cache.Cache
}

func NewIndicatorService(repo *repository.IndicatorRepository, c *cache.Cache) *IndicatorService {
	return &IndicatorService{
		repo:  repo,
		cache: c,
	}
}

func (s *IndicatorService) GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error) {
	cacheKey := cache.GenerateKey("indicator", map[string]string{"id": id})
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.IndicatorWithRelations), nil
	}

	indicator, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, indicator, cache.TTLIndicatorDetail)
	return indicator, nil
}

func (s *IndicatorService) Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	cacheKey := cache.GenerateKey("search", params)
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.SearchResult), nil
	}

	result, err := s.repo.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, result, cache.TTLIndicatorSearch)
	return result, nil
}
