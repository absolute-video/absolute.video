package types


// Stream struct only includes the height and width
type Stream struct {
    Height int `json:"height"`
    Width  int `json:"width"`
}

// Wrapper struct that includes a slice of Stream
type VideoData struct {
    Streams []Stream `json:"streams"`
}

type TranscodeSize struct {
	Height int
	Width int
}
