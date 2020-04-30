package model

type GeoLiteCityBlocks struct {
	ID          int64 //ID
	StartIP     string
	EndIP       string
	StartIPNum  int64
	EndIPNum    int64
	Address     string
	NodeAddress string
	LocalId     int64
	Country     string
	Province    string
	City        string
	//GeoLiteCityLocation GeoLiteCityLocation `gorm:"ForeignKey:LocalId"`
}

type Cities struct {
	ID       int64
	Country  string
	Province string
	City     string
}

type Provinces struct {
	ID       int64
	Country  string
	Province string
}

type Country struct {
	ID      int64
	Country string
}
