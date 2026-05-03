package service

import (
	"context"
	"time"

	"dooz/entity"
	"dooz/internal/app/api/dto"
	storeItemRepo "dooz/internal/repository/store_item"
	userRepo "dooz/internal/repository/user"
	userItemRepo "dooz/internal/repository/user_item"
)

type ShopService interface {
	GetItems(ctx context.Context) ([]*dto.StoreItemDTO, error)
	GetItemByKey(ctx context.Context, itemKey string) (*dto.StoreItemDTO, error)
	GetInventory(ctx context.Context, userID string) (*dto.InventoryDTO, error)
	Purchase(ctx context.Context, userID string, itemType byte, itemValue int) error
	UpdateCurrentStyle(ctx context.Context, userID string, req *dto.UpdateCurrentStyleRequest) error
	AddWallet(ctx context.Context, userID string, coins, gems int) error
	GrantDefaultItems(ctx context.Context, userID string) error
}

type shopService struct {
	userRepo      userRepo.Repository
	storeItemRepo storeItemRepo.Repository
	userItemRepo  userItemRepo.Repository
}

func NewShopService(
	userRepo userRepo.Repository,
	storeItemRepo storeItemRepo.Repository,
	userItemRepo userItemRepo.Repository,
) ShopService {
	return &shopService{
		userRepo:      userRepo,
		storeItemRepo: storeItemRepo,
		userItemRepo:  userItemRepo,
	}
}

func (s *shopService) GetItems(ctx context.Context) ([]*dto.StoreItemDTO, error) {
	items, err := s.storeItemRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.StoreItemDTO, len(items))
	for i, item := range items {
		out[i] = &dto.StoreItemDTO{
			ID:            item.ID,
			ItemType:      int(item.ItemType),
			ItemValue:     item.ItemValue,
			ItemKey:       item.ItemKey,
			AssetURL:      item.AssetURL,
			PriceCurrency: int(item.PriceCurrency),
			PriceAmount:   item.PriceAmount,
		}
	}
	return out, nil
}

func (s *shopService) GetItemByKey(ctx context.Context, itemKey string) (*dto.StoreItemDTO, error) {
	item, err := s.storeItemRepo.GetByKey(ctx, itemKey)
	if err != nil {
		return nil, err
	}
	return &dto.StoreItemDTO{
		ID:            item.ID,
		ItemType:      int(item.ItemType),
		ItemValue:     item.ItemValue,
		ItemKey:       item.ItemKey,
		AssetURL:      item.AssetURL,
		PriceCurrency: int(item.PriceCurrency),
		PriceAmount:   item.PriceAmount,
	}, nil
}

func (s *shopService) GetInventory(ctx context.Context, userID string) (*dto.InventoryDTO, error) {
	userItems, err := s.userItemRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	catalog, err := s.storeItemRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	catalogByID := make(map[string]*entity.StoreItem, len(catalog))
	for _, item := range catalog {
		catalogByID[item.ID] = item
	}

	inv := &dto.InventoryDTO{
		Themes:   []dto.OwnedItemDTO{},
		XOShapes: []dto.OwnedItemDTO{},
		Avatars:  []dto.OwnedItemDTO{},
	}
	for _, owned := range userItems {
		item, ok := catalogByID[owned.ItemID]
		if !ok {
			continue
		}
		switch item.ItemType {
		case entity.StoreItemTheme:
			inv.Themes = append(inv.Themes, dto.OwnedItemDTO{
				ItemValue: item.ItemValue,
				AssetURL:  item.AssetURL,
			})
		case entity.StoreItemXOShape:
			inv.XOShapes = append(inv.XOShapes, dto.OwnedItemDTO{
				ItemValue: item.ItemValue,
				AssetURL:  item.AssetURL,
			})
		case entity.StoreItemAvatar:
			inv.Avatars = append(inv.Avatars, dto.OwnedItemDTO{
				ItemValue: item.ItemValue,
				AssetURL:  item.AssetURL,
			})
		}
	}
	return inv, nil
}

func (s *shopService) Purchase(ctx context.Context, userID string, itemType byte, itemValue int) error {
	item, err := s.storeItemRepo.GetByTypeAndValue(ctx, itemType, itemValue)
	if err != nil {
		return err
	}
	hasItem, err := s.userItemRepo.HasItem(ctx, userID, item.ID)
	if err != nil {
		return err
	}
	if hasItem {
		return userItemRepo.ErrAlreadyOwned
	}

	if item.PriceAmount > 0 {
		if err := s.userRepo.DeductBalance(ctx, userID, byte(item.PriceCurrency), item.PriceAmount); err != nil {
			return err
		}
	}

	return s.userItemRepo.Create(ctx, &entity.UserItem{
		UserID:     userID,
		ItemID:     item.ID,
		AcquiredAt: time.Now().Unix(),
	})
}

func (s *shopService) UpdateCurrentStyle(ctx context.Context, userID string, req *dto.UpdateCurrentStyleRequest) error {
	if req.CurrentTheme != nil {
		if err := s.ensureOwnedByTypeValue(ctx, userID, byte(entity.StoreItemTheme), *req.CurrentTheme); err != nil {
			return err
		}
	}
	if req.CurrentXOShape != nil {
		if err := s.ensureOwnedByTypeValue(ctx, userID, byte(entity.StoreItemXOShape), *req.CurrentXOShape); err != nil {
			return err
		}
	}
	if req.CurrentAvatar != nil {
		if err := s.ensureOwnedByTypeValue(ctx, userID, byte(entity.StoreItemAvatar), *req.CurrentAvatar); err != nil {
			return err
		}
	}
	return s.userRepo.UpdateCurrentStyle(ctx, userID, req.CurrentTheme, req.CurrentXOShape, req.CurrentAvatar)
}

func (s *shopService) ensureOwnedByTypeValue(ctx context.Context, userID string, itemType byte, itemValue int) error {
	item, err := s.storeItemRepo.GetByTypeAndValue(ctx, itemType, itemValue)
	if err != nil {
		return err
	}
	hasItem, err := s.userItemRepo.HasItem(ctx, userID, item.ID)
	if err != nil {
		return err
	}
	if !hasItem {
		return userItemRepo.ErrNotOwned
	}
	return nil
}

func (s *shopService) AddWallet(ctx context.Context, userID string, coins, gems int) error {
	return s.userRepo.AddBalance(ctx, userID, coins, gems)
}

func (s *shopService) GrantDefaultItems(ctx context.Context, userID string) error {
	defaultItems := []struct {
		itemType  byte
		itemValue int
	}{
		{byte(entity.StoreItemTheme), 1},
		{byte(entity.StoreItemXOShape), 1},
		{byte(entity.StoreItemAvatar), 1},
	}
	now := time.Now().Unix()
	for _, defaultItem := range defaultItems {
		item, err := s.storeItemRepo.GetByTypeAndValue(ctx, defaultItem.itemType, defaultItem.itemValue)
		if err != nil {
			return err
		}
		hasItem, err := s.userItemRepo.HasItem(ctx, userID, item.ID)
		if err != nil {
			return err
		}
		if hasItem {
			continue
		}
		if err := s.userItemRepo.Create(ctx, &entity.UserItem{
			UserID:     userID,
			ItemID:     item.ID,
			AcquiredAt: now,
		}); err != nil {
			return err
		}
	}
	return nil
}
