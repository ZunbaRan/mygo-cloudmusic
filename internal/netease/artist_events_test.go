package netease

import "testing"

// An event's content comes as a string of JSON, sharing a song in the old
// shape, or passing on another event, whose words show with its poster's.
func TestDecodeEvent(t *testing.T) {
	c := decodeEvent(`{"msg":"New Songs","song":{"id":1,"name":"演员","artists":[{"id":5781,"name":"Joker Xue"}],"album":{"id":2,"name":"绅士","picUrl":"p"},"duration":261000}}`)
	if c.Msg != "New Songs" || c.Song == nil || c.Song.ArtistNames() != "Joker Xue" || c.Song.Album.ID != 2 || c.Song.Duration != 261000 {
		t.Fatalf("song: %+v %+v", c, c.Song)
	}
	c = decodeEvent(`{"msg":"转发","event":{"user":{"nickname":"Someone"},"json":"{\"msg\":\"原话\"}"}}`)
	if c.Forwarded != "Someone：原话" {
		t.Fatalf("forwarded %q", c.Forwarded)
	}
	if c := decodeEvent(`{"msg":"看","mv":{"id":3}}`); !c.Video {
		t.Fatal("an MV isn't a video")
	}
}
