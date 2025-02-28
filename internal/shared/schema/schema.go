package schema

type Endpoint struct {
	ID   uint `gorm:"primaryKey"`
	IP   string
	Port string
}

type Rule struct {
	ID         uint   `gorm:"primaryKey"` // Syn with name, name should be unique
	Program    string // path to the program
	Protocol   string // protocol (TCP, UDP, ICMP, Any)
	RemoteAddr string // remote IP address
	Action     string // block
	Enabled    bool   // enable or disable the rule // Can be yes or no
}

type EndpointRule struct {
	EndpointID uint `gorm:"primaryKey"`
	RuleID     uint `gorm:"primaryKey"`
}
