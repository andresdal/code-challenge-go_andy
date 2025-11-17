package models

type Book struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Author    string `json:"author"`
	UnitsSold uint   `json:"units_sold"`
	Price     uint   `json:"price"`
}
