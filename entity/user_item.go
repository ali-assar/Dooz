package entity

type UserItem struct {
	ID         string `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID     string `gorm:"column:user_id;type:uuid;not null;index" json:"user_id"`
	ItemID     string `gorm:"column:item_id;type:uuid;not null;index" json:"item_id"`
	AcquiredAt int64  `gorm:"column:acquired_at;type:bigint;not null" json:"acquired_at"`
}

func (UserItem) TableName() string {
	return "user_items"
}
