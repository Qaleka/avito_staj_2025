package domain

import "context"

var MerchTypes = map[string]int{
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}

type Inventory struct {
	ID         int    `gorm:"primary_key;auto_increment;column:id" json:"id"`
	OwnerID    string `gorm:"column:owner_id;not null;index:idx_owner_item,unique" json:"ownerID"`
	ItemName   string `gorm:"type:varchar(255);column:item_name;not null;index:idx_owner_item,unique" json:"itemName"`
	ItemAmount int    `gorm:"column:item_amount;not null" json:"itemAmount"`
	User       User   `gorm:"foreignkey:OwnerID;references:UUID" json:"-"`
}

type Transaction struct {
	UUID       string `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:uuid" json:"id"`
	SenderID   string `gorm:"column:sender_id;not null" json:"senderID"`
	ReceiverID string `gorm:"column:receiver_id;not null" json:"receiverID"`
	Amount     int    `gorm:"type:int;column:amount;not null" json:"amount"`
	Sender     User   `gorm:"foreignkey:SenderID;references:UUID" json:"-"`
	Receiver   User   `gorm:"foreignkey:ReceiverID;references:UUID" json:"-"`
}

type TransactionWithUsers struct {
	SenderName   string
	ReceiverName string
	Amount       int
}

type UserInformationResponse struct {
	Coins       int                 `gorm:"type:int;default:0;column:coins" json:"coins"`
	Inventory   []InventoryResponse `gorm:"foreignkey:InventoryID;references:ID" json:"inventory"`
	CoinHistory CoinHistory         `gorm:"foreignkey:CoinHistoryID;references:ID" json:"coinHistory"`
}

type InventoryResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []ReceivedResponse `json:"received"`
	Sent     []SentResponse     `json:"sent"`
}

type ReceivedResponse struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentResponse struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SentRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type MerchRepository interface {
	GetUserMerchInformation(ctx context.Context, userID string) (UserInformationResponse, error)
	SendCoins(ctx context.Context, senderID string, receiverID string, amount int) error
	BuyItem(ctx context.Context, userID string, itemName string, itemCost int) error
}
