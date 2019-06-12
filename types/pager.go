package types

// Pager represent a generic paging parameter
type Pager interface {
	Offset() int
	Limit() int
}
