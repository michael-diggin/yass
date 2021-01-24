package convert

import (
	"encoding/json"

	"github.com/michael-diggin/yass/common/models"
	"github.com/pkg/errors"

	pb "github.com/michael-diggin/yass/proto"
)

// ToModel converts the proto type to a model type
func ToModel(p *pb.Pair) (*models.Pair, error) {
	var val interface{}
	if p.Value != nil {
		err := json.Unmarshal(p.Value, &val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal value")
		}
	}
	return &models.Pair{Key: p.Key, Hash: p.Hash, Value: val}, nil
}

// ToPair converts the model type to a proto type
func ToPair(p *models.Pair) (*pb.Pair, error) {
	var val = []byte{}
	var err error
	if p.Value != nil {
		val, err = json.Marshal(p.Value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal value")
		}
	}
	return &pb.Pair{Key: p.Key, Hash: p.Hash, Value: val}, nil
}
