package tools

import (
	"encoding/json"

	"github.com/sc-js/backend_core/src/mongowrap"
	"github.com/speps/go-hashids"
	"gorm.io/gorm"
)

type ModelID uint

func (i ModelID) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

var h *hashids.HashID

func Init(salt string) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 10
	if h == nil {
		h, _ = hashids.NewWithData(hd)
	}
}

func (i ModelID) MarshalJSON() ([]byte, error) {
	e := ""

	if i > 0 {
		e, _ = h.Encode([]int{int(i)})
	}
	return json.Marshal(e)
}

func Decode(hid string) ModelID {
	d, err := h.DecodeWithError(hid)
	if err != nil {
		return 0
	}
	return ModelID(d[0])
}

func Encode(id ModelID) string {
	e, err := h.Encode([]int{int(id)})
	if err == nil {
		return e
	}
	return ""
}

func (i *ModelID) UnmarshalJSON(data []byte) error {
	tmp := ""
	if err := json.Unmarshal(data, &tmp); err != nil {
		tmpUint := uint(0)
		if err := json.Unmarshal(data, &tmpUint); err == nil {
			tmpModelId := ModelID(tmpUint)
			*i = tmpModelId
			return nil
		}
		return err
	}
	d, _ := h.DecodeWithError(tmp)

	if len(d) > 0 {
		if d[0] > 0 {
			tmp := ModelID(d[0])
			*i = tmp
		}
	}

	return nil
}

type DataWrap struct {
	DB    *gorm.DB
	Mongo *mongowrap.Mongo
}
