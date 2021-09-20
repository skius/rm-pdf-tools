package actions

type Action interface {
	Page() int
}

type T []Action

type Delete struct {
	Count int
	PageNo int
}
func (d Delete) Page() int {
	return d.PageNo
}

type Insert struct {
	Count int
	PageNo int
	InsertAfter bool
}
func (i Insert) Page() int {
	return i.PageNo
}

type PageReplacement struct {
	OriginalIdx int
	NewIdx int
	Deleted bool
}

