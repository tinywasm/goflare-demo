package contact

import (
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/json"
	"github.com/tinywasm/orm"
)

func HandleList(db *orm.DB) router.HandlerFunc {
	return func(ctx router.Context) {
		ctx.SetHeader("Content-Type", "application/json")
		ctx.SetHeader("Access-Control-Allow-Origin", "*")

		qb := db.Query(&Contact{}).OrderBy("id").Desc()
		list, err := ReadAllContact(qb)
		if err != nil {
			ctx.WriteStatus(502)
			ctx.Write([]byte(`{"error":"db error"}`))
			return
		}
		// json.Encode(data fmt.Fielder, output any) — output: *[]byte | *string | io.Writer.
		// ContactList implementa fmt.FielderSlice → se serializa como array.
		var body []byte
		if err := json.Encode(list, &body); err != nil {
			ctx.WriteStatus(500)
			ctx.Write([]byte(`{"error":"encode error"}`))
			return
		}
		ctx.WriteStatus(200)
		ctx.Write(body)
	}
}
