package model

import (
	"errors"
	"gorm.io/gorm"
)

type User struct {
	Uid int64 `gorm:"primaryKey" json:"uid"`
	UserName string `gorm:"column:user_name" json:"user_name"`
	UserPassword string `gorm:"column:user_password" json:"user_password"`
}

func (u *User) create() error {
	return DB.Create(u).Error
}

func (u *User) modify(u2 *User) error {
	u2.Uid = u.Uid
	return DB.Where("uid = ?",u.Uid).Save(u2).Error
}

func (u *User) delete() error{
	return DB.Where("uid = ?",u.Uid).Delete(u).Error
}

func (u *User) uidSelect(uid int64) error{
	return DB.First(u,uid).Error
}

func (u *User) userNameSelect(userName string) error {
	return DB.First(u,"user_name = ?",userName).Error
}

func (u *User) SelectUser1(uid int64) error{
	return u.uidSelect(uid)
}

func (u *User) SelectUser2(userName string) error {
	return u.userNameSelect(userName)
}

func (u *User) VerifyPassword(userName string,password string) (*User,error) {
	err := DB.Model(u).Where("user_name = ? and user_password = ?",userName,password).First(&u).Error
	if errors.Is(err,gorm.ErrRecordNotFound) {
		return nil,errors.New("用户不存在")
	} else if err != nil {
		return nil,err
	}
	return u,nil
}

func (u *User) CreateUser(user *User) error {
	err := u.userNameSelect(user.UserName)
	if err != nil {
		if errors.Is(err,gorm.ErrRecordNotFound) {
			return user.create()
		} else {
			return err
		}
	} else {
		return errors.New("用户名已存在")
	}
}

func (u *User) ModifyUser(user *User) error {
	err := user.modify(user)
	if errors.Is(err,gorm.ErrRecordNotFound) {
		return errors.New("用户不存在")
	}
	return err
}

func (u *User) DeleteUser(user *User) error {
	err := user.delete()
	if errors.Is(err,gorm.ErrRecordNotFound) {
		return errors.New("用户不存在")
	}
	return err
}