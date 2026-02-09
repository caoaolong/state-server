package orm

import "encoding/json"

func CreateFlow(data string) error {
	json.Unmarshal([]byte(data), &SMFlow{})
	return db.Create(&SMFlow{}).Error
}
