package model

type UserRecord struct {
	Uid int64 `gorm:"column:uid"`
	VideoId string `gorm:"column:video_id"`
}

func (ur *UserRecord) SelectAll(uid int64) ([]UserRecord,error) {
	var urs []UserRecord
	err := DB.Find(&urs,"uid = ?",uid).Error
	return urs,err
}

func (ur *UserRecord) Delete() error {
	return DB.Delete(ur).Error
}

func (ur *UserRecord) Create() error {
	return DB.Create(ur).Error
}