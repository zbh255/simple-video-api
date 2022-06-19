package model

import (
	"fmt"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"testing"
)

func TestUser(t *testing.T) {
	InitOrm("../testdata/test.db")
	u := &User{}
	u.UserName = "Tony"
	u.UserPassword = "123456"
	err := u.CreateUser(u)
	if err != nil {
		t.Fatal(err)
	}
	err = u.ModifyUser(&User{UserName: "Jeni",UserPassword: "1234"})
	if err != nil {
		t.Fatal(err)
	}
	err = u.DeleteUser(&User{Uid: 0,UserName: "Jeni"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUserRecord(t *testing.T) {
	InitOrm("../testdata/test.db")
	users := make([]User,0)
	for i := 0; i < 10; i++ {
		user := User{}
		user.UserName = fmt.Sprintf("man%d",i)
		user.UserPassword = strconv.Itoa(877 + i)
		users = append(users,user)
	}
	for _,v := range users {
		err := v.CreateUser(&v)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < int(rand.Int31n(100)); i++ {
			record := UserRecord{}
			record.Uid = v.Uid
			record.VideoId = uuid.New().String()
			err := record.Create()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	record := &UserRecord{}
	records,err := record.SelectAll(1)
	if err != nil {
		t.Fatal(err)
	}
	for _,v := range records {
		t.Log(v.VideoId)
	}
	for _,v := range users {
		err := v.SelectUser2(v.UserName)
		if err != nil {
			t.Fatal(err)
		}
		err = v.DeleteUser(&v)
		if err != nil {
			t.Fatal(err)
		}
	}
}
