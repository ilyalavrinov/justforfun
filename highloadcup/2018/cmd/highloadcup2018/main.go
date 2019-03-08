package main

import (
	"fmt"
	"github.com/admirallarimda/highloadcup2018/internal/pkg/hlcup"
	hlredis "github.com/admirallarimda/highloadcup2018/internal/pkg/hlcup/redis"
	"github.com/go-redis/redis"
	"github.com/gramework/gramework"
	"strings"
)

var filterPredicates map[string]map[string]bool

func init() {
	filterPredicates = make(map[string]map[string]bool, 15)
	filterPredicates["sex"] = map[string]bool{"eq": true}
	filterPredicates["email"] = map[string]bool{"domain": true, "lt": true, "gt": true}
	filterPredicates["status"] = map[string]bool{"eq": true, "neq": true}
	filterPredicates["fname"] = map[string]bool{"eq": true, "any": true, "null": true}
	filterPredicates["sname"] = map[string]bool{"eq": true, "starts": true, "null": true}
	filterPredicates["phone"] = map[string]bool{"code": true, "null": true}
	filterPredicates["country"] = map[string]bool{"eq": true, "null": true}
	filterPredicates["city"] = map[string]bool{"eq": true, "any": true, "null": true}
	filterPredicates["birth"] = map[string]bool{"lt": true, "gt": true, "year": true}
	filterPredicates["interests"] = map[string]bool{"contains": true, "any": true}
	filterPredicates["likes"] = map[string]bool{"contains": true}
	filterPredicates["premium"] = map[string]bool{"now": true, "null": true}
}

func main() {

	storage := hlredis.NewRedisAccountStorage(redis.Options{Addr: "localhost:6379"})

	app := gramework.New()
	app.POST("/accounts/new/", func(ctx *gramework.Context) {
		handleNewAccount(ctx, storage)
	})
	app.GET("/accounts/filter/", func(ctx *gramework.Context) {
		handleFilter(ctx, storage)
	})
	app.ListenAndServe()
}

func handleNewAccount(ctx *gramework.Context, saver hlcup.AccountSaver) {
	//ctx.Logger.Debugf("Incomming msg on new account: %s", ctx.Request.Body())

	acc := hlcup.RawAccount{}
	if err := ctx.UnJSON(&acc); err != nil {
		ctx.Err500("Could not parse account json: ", err)
		return
	}

	//ctx.Logger.Debugf("Incoming account: %+v", acc)

	if err := saver.Save(acc); err != nil {
		ctx.Err500("Could not save account into storage: ", err)
		return
	}
}

func handleFilter(ctx *gramework.Context, filter hlcup.AccountFilter) {
	filterset := hlcup.FilterSet{}
	for k, v := range ctx.GETParams() {
		if k == "query_id" {
			// just a service parameter
			continue
		}

		keyparts := strings.Split(k, "_")
		if len(keyparts) != 2 {
			ctx.BadRequest(fmt.Errorf("Filter key '%s' could not be split", k))
			return
		}
		field := keyparts[0]
		pred := keyparts[1]

		preds, found := filterPredicates[field]
		if !found {
			ctx.BadRequest(fmt.Errorf("Filter field '%s' is unknown", field))
			return
		}
		if preds[pred] == false {
			ctx.BadRequest(fmt.Errorf("Filter predicate '%s' cannot be used for field '%s'", pred, field))
			return
		}

		switch field {
		case "sex":
			filterset.Sex = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "email":
			filterset.Email = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "status":
			filterset.Status = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "fname":
			filterset.Firstname = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "sname":
			filterset.Surname = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "phone":
			filterset.Phone = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "country":
			filterset.Country = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "city":
			filterset.City = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "birth":
			filterset.Birth = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "interests":
			filterset.Interests = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "likes":
			filterset.Likes = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		case "premium":
			filterset.Premium = &hlcup.Filter{Type: hlcup.FilterType(pred), Values: v}
		}
	}
}
