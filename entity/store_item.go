package entity

type StoreItemType byte
type CurrencyType byte

const (
	StoreItemTheme   StoreItemType = 1
	StoreItemXOShape StoreItemType = 2
	StoreItemAvatar  StoreItemType = 3
)

const (
	CurrencyCoins CurrencyType = 1
	CurrencyGems  CurrencyType = 2
)

type StoreItem struct {
	ID            string        `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ItemType      StoreItemType `gorm:"column:item_type;type:smallint;not null" json:"item_type"`
	ItemValue     int           `gorm:"column:item_value;type:smallint;not null" json:"item_value"`
	ItemKey       string        `gorm:"column:item_key;type:text;uniqueIndex;not null" json:"item_key"`
	AssetURL      string        `gorm:"column:asset_url;type:text;not null;default:''" json:"asset_url"`
	PriceCurrency CurrencyType  `gorm:"column:price_currency;type:smallint;not null" json:"price_currency"`
	PriceAmount   int           `gorm:"column:price_amount;type:integer;not null;default:0" json:"price_amount"`
	IsActive      bool          `gorm:"column:is_active;type:boolean;not null;default:true" json:"is_active"`
}

func (StoreItem) TableName() string {
	return "store_items"
}
