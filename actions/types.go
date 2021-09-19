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

// Content represents the top-level UUID.content JSON
type Content struct {
	CoverPageNumber  int `json:"coverPageNumber"`
	DocumentMetadata struct {
	} `json:"documentMetadata"`
	DummyDocument bool `json:"dummyDocument"`
	ExtraMetadata struct {
		LastBallpointv2Color   string `json:"LastBallpointv2Color"`
		LastBallpointv2Size    string `json:"LastBallpointv2Size"`
		LastCalligraphyColor   string `json:"LastCalligraphyColor"`
		LastCalligraphySize    string `json:"LastCalligraphySize"`
		LastEraseSectionColor  string `json:"LastEraseSectionColor"`
		LastEraseSectionSize   string `json:"LastEraseSectionSize"`
		LastEraserColor        string `json:"LastEraserColor"`
		LastEraserSize         string `json:"LastEraserSize"`
		LastEraserTool         string `json:"LastEraserTool"`
		LastFinelinerv2Color   string `json:"LastFinelinerv2Color"`
		LastFinelinerv2Size    string `json:"LastFinelinerv2Size"`
		LastHighlighterv2Color string `json:"LastHighlighterv2Color"`
		LastHighlighterv2Size  string `json:"LastHighlighterv2Size"`
		LastPaintbrushv2Color  string `json:"LastPaintbrushv2Color"`
		LastPaintbrushv2Size   string `json:"LastPaintbrushv2Size"`
		LastPen                string `json:"LastPen"`
		LastPencilv2Color      string `json:"LastPencilv2Color"`
		LastPencilv2Size       string `json:"LastPencilv2Size"`
		LastSelectionToolColor string `json:"LastSelectionToolColor"`
		LastSelectionToolSize  string `json:"LastSelectionToolSize"`
		LastSharpPencilv2Color string `json:"LastSharpPencilv2Color"`
		LastSharpPencilv2Size  string `json:"LastSharpPencilv2Size"`
		LastTool               string `json:"LastTool"`
		LastUndefinedColor     string `json:"LastUndefinedColor"`
		LastUndefinedSize      string `json:"LastUndefinedSize"`
	} `json:"extraMetadata"`
	FileType      string   `json:"fileType"`
	FontName      string   `json:"fontName"`
	LineHeight    int      `json:"lineHeight"`
	Margins       int      `json:"margins"`
	Orientation   string   `json:"orientation"`
	PageCount     int      `json:"pageCount"`
	Pages         []string `json:"pages"`
	TextAlignment string   `json:"textAlignment"`
	TextScale     int      `json:"textScale"`
}
