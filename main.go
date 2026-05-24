package main

import (
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"rinha2026/model"
	"rinha2026/normalize"
	"rinha2026/vptree"

	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
)

const k = 5

var (
	fraudTree  *vptree.VPTree
	normParams model.Params
	mccRisk    map[string]float64
)

func init() {
	var err error

	fraudTree, err = vptree.Load("resources/vptree.bin")
	if err != nil {
		panic(err)
	}

	normParams, err = normalize.LoadParams("resources/normalization.json")
	if err != nil {
		panic(err)
	}

	mccRisk, err = normalize.LoadMccRisk("resources/mcc_risk.json")
	if err != nil {
		panic(err)
	}
}

func main() {
	readyFn := func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetStatusCode(200)
	}

	fraudScoreFn := func(ctx *fasthttp.RequestCtx) {
		var req model.FraudScoreRequest
		if err := json.Unmarshal(ctx.Request.Body(), &req); err != nil {
			ctx.Error("error parsing request", http.StatusBadRequest)
			return
		}

		vector := normalize.ToVectorQuantized(req, normParams, mccRisk)
		neighbors := vptree.Search(fraudTree, vector)
		fraudScore := computeFraudScore(neighbors)
		approved := fraudScore < 0.5

		res := model.FraudScoreResponse{
			Approved:   approved,
			FraudScore: fraudScore,
		}

		ctx.Response.Header.SetContentType("application/json")
		data, err := json.Marshal(res)
		if err != nil {
			ctx.Error("error parsing response", http.StatusBadRequest)
		}
		ctx.Response.SetBody(data)
	}

	handler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		switch path {
		case "/fraud-score":
			fraudScoreFn(ctx)

		case "/ready":
			readyFn(ctx)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	log.Printf("Starting server on port %s...", port)
	if err := fasthttp.ListenAndServe(":"+port, handler); err != nil {
		panic(err)
	}
}

func computeFraudScore(neighbors [5]int) float64 {
	fraudCount := 0
	for i := range k {
		idx := neighbors[i]
		if idx < 0 {
			break
		}
		if !fraudTree.Nodes[idx].Label {
			fraudCount++
		}
	}
	if fraudCount == 0 {
		return 0
	}
	return float64(fraudCount) / float64(k)
}
