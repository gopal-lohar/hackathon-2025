package store

import (
	"strconv"

	"github.com/gopal-lohar/hackathon-2025/internal/shared/schema"
	"gorm.io/gorm"
)

type RealRule struct {
	ID         uint   `json:"id"`
	EndpointID string `json:"endpoint_id"`
	Program    string `json:"program"`
	Protocol   string `json:"protocol"`
	RemoteIP   string `json:"remote_ip"`
	Action     string `json:"action"`
	Enabled    bool   `json:"enabled"`
}

type RuleStore struct {
	DB *gorm.DB
}

func NewRuleStore(db *gorm.DB) *RuleStore {
	return &RuleStore{DB: db}
}

// Name    		 string // unique name of the rule (if not unique the command netsh will fail)
// Program    string // path to the program
// Protocol   string // protocol (TCP, UDP, ICMP, Any)
// RemoteAddr string // remote IP address
// Action     string // block
// Enabled    bool   // enable or disable the rule
type Temp struct {
	EndpointID string
	Enabled    bool
}

func (rs *RuleStore) AddRule(program, protocol, remoteAddr, action string, enabled bool, temp Temp) (int, error) {
	existingRule := &schema.Rule{}
	err := rs.DB.Where("program = ? AND protocol = ? AND remote_addr = ? AND action = ?", program, protocol, remoteAddr, action).First(existingRule).Error
	if err == nil {
		return -1, nil
	}
	rule := &schema.Rule{
		Program:    program,
		Protocol:   protocol,
		RemoteAddr: remoteAddr,
		Action:     action,
	}
	err = rs.DB.Create(rule).Error
	if err != nil {
		return 0, err
	}

	// Search the rule to return its id
	searchRule := &schema.Rule{}

	err = rs.DB.Where("program = ? AND protocol = ? AND remote_addr = ? AND action = ?", program, protocol, remoteAddr, action).First(searchRule).Error
	if err != nil {
		return 0, err
	}

	if temp.Enabled == true {
		endpointidInt, _ := strconv.Atoi(temp.EndpointID)
		endpointRule := &schema.EndpointRule{
			EndpointID: uint(endpointidInt),
			RuleID:     searchRule.ID,
		}
		err = rs.DB.Create(endpointRule).Error
		if err != nil {
			return 0, err
		}
	}
	return int(searchRule.ID), nil
}

func (rs *RuleStore) DeleteRule(name string) error {
	return rs.DB.Where("id = ?", name).Delete(&schema.Rule{}).Error
}

func (rs *RuleStore) GetRules() ([]RealRule, error) {
	rules := []schema.Rule{}
	err := rs.DB.Find(&rules).Error

	realRules := make([]RealRule, 0)
	for _, rule := range rules {
		endpointId := rs.GetEndpointId(rule.ID)
		realRules = append(realRules, RealRule{
			ID:         rule.ID,
			EndpointID: endpointId,
			Program:    rule.Program,
			Protocol:   rule.Protocol,
			RemoteIP:   rule.RemoteAddr,
			Action:     rule.Action,
			Enabled:    rule.Enabled,
		})
	}
	return realRules, err
}

func (rs *RuleStore) GetEndpointId(ruleId uint) string {
	endpointRule := &schema.EndpointRule{}
	err := rs.DB.Where("rule_id = ?", ruleId).First(endpointRule).Error
	if err != nil {
		return ""
	}
	return strconv.Itoa(int(endpointRule.EndpointID))
}
