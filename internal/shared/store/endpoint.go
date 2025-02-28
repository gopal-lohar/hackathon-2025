package store

import (
	"errors"
	"strconv"

	"github.com/rudransh-shrivastava/context-aware-firewall/internal/shared/schema"
	"gorm.io/gorm"
)

type EndpointStore struct {
	DB *gorm.DB
}

func NewEndpointStore(db *gorm.DB) *EndpointStore {
	return &EndpointStore{DB: db}
}

// returns endpointID and error
func (es *EndpointStore) CreateEndpoint(ip string, port string) (string, error) {
	endpoint := &schema.Endpoint{
		IP:   ip,
		Port: port,
	}
	err := es.DB.Create(endpoint).Error
	if err != nil {
		return "", err
	}
	searchedEndpoint := &schema.Endpoint{}
	err = es.DB.Where("ip = ? AND port = ?", ip, port).First(searchedEndpoint).Error
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(searchedEndpoint.ID)), nil
}


func (es *EndpointStore) GetEndpoints() ([]schema.Endpoint, error) {
	var endpoints []schema.Endpoint
	err := es.DB.Find(&endpoints).Error
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}


func (es *EndpointStore) DeleteEndpoint(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	result := es.DB.Delete(&schema.Endpoint{}, uint(idInt))
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("row does not exit")
	}

	return nil
}