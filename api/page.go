package api

import (
	"strconv"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/types"
)

var (
	defaultPageLimit = 20
)

var (
	defaultPageParam = &pageParam{
		limit:  defaultPageLimit,
		offset: 0,
	}
)

type pageParam struct {
	limit  int
	offset int
}

func (p *pageParam) Offset() int { return p.offset }
func (p *pageParam) Limit() int  { return p.limit }

func getPager(ctx *httpmux.Context) types.Pager {
	var (
		limit  = ctx.Query["limit"]
		offset = ctx.Query["offset"]
	)
	if limit == "" && offset == "" {
		return defaultPageParam
	}

	limitN, err := strconv.Atoi(limit) // invalid/negative=20
	if err != nil || limitN < 0 {
		limitN = defaultPageLimit
	}

	offsetN, _ := strconv.Atoi(offset) // invalid/negative=0
	if offsetN < 0 {
		offsetN = 0
	}

	return &pageParam{limitN, offsetN}
}
