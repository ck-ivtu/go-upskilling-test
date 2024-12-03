package model

type DeliveryPackage struct {
	ID              string `gorm:"primary_key" json:"id"`
	CustomerEmail   string `gorm:"column:delivery_address" json:"customer_email"`
	DeliveryAddress string `gorm:"column:customer_email" json:"delivery_address"`
}
