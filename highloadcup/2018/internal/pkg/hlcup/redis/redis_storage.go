package redis

import (
	"fmt"

	"github.com/admirallarimda/highloadcup2018/internal/pkg/hlcup"
	"github.com/go-redis/redis"
)

type redisStorage struct {
	client *redis.Client
}

func NewRedisAccountStorage(opts redis.Options) hlcup.AccountStorage {
	client := redis.NewClient(&opts)
	return &redisStorage{client}
}

func (r *redisStorage) Save(account hlcup.RawAccount) error {
	fields := packAccount(account)

	return r.client.HMSet(keyAccount(account.ID), fields).Err()
}

func keyAccount(id int32) string {
	return fmt.Sprintf("acc:%d", id)
}

func packAccount(a hlcup.RawAccount) map[string]interface{} {
	fields := make(map[string]interface{}, 20)

	fields["id"] = a.ID
	fields["email"] = a.EMail
	fields["fname"] = a.Firstname
	fields["sname"] = a.Surname
	fields["phone"] = a.Phone
	fields["sex"] = a.Sex
	fields["birth"] = a.BirthTimestamp
	fields["country"] = a.Country
	fields["city"] = a.City

	fields["joined"] = a.JoinedTimestamp
	fields["status"] = string(a.Status)
	// TODO:
	// interests, premium, likes

	return fields
}

func unpackAccount(fields map[string]interface{}) hlcup.RawAccount {
	return hlcup.RawAccount{
		ID:             fields["id"].(int32),
		EMail:          fields["email"].(string),
		Firstname:      fields["fname"].(string),
		Surname:        fields["sname"].(string),
		Phone:          fields["phone"].(string),
		Sex:            fields["sex"].(string),
		BirthTimestamp: fields["birth"].(int64),
		Country:        fields["country"].(string),
		City:           fields["city"].(string),

		JoinedTimestamp: fields["joined"].(int64),
		Status:          fields["status"].(hlcup.MaritalStatus)}
}

func (r *redisStorage) Filter(filter hlcup.FilterSet) []int32 {
	return nil
}
