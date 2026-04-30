package dto

type StoreItemDTO struct {
	ID            string `json:"id"`
	ItemType      int    `json:"item_type"`
	ItemValue     int    `json:"item_value"`
	ItemKey       string `json:"item_key"`
	AssetURL      string `json:"asset_url"`
	PriceCurrency int    `json:"price_currency"`
	PriceAmount   int    `json:"price_amount"`
}

type OwnedItemDTO struct {
	ItemValue int    `json:"item_value"`
	AssetURL  string `json:"asset_url"`
}

type InventoryDTO struct {
	Themes   []OwnedItemDTO `json:"themes"`
	XOShapes []OwnedItemDTO `json:"xo_shapes"`
	Avatars  []OwnedItemDTO `json:"avatars"`
}

type PurchaseItemRequest struct {
	ItemType  int `json:"item_type" binding:"required"`
	ItemValue int `json:"item_value" binding:"required"`
}

type UpdateCurrentStyleRequest struct {
	CurrentTheme   *int `json:"current_theme,omitempty"`
	CurrentXOShape *int `json:"current_xo_shape,omitempty"`
	CurrentAvatar  *int `json:"current_avatar,omitempty"`
}

type AddWalletRequest struct {
	Coins int `json:"coins"`
	Gems  int `json:"gems"`
}
