package ytfeed

import (
	"context"
	"encoding/json"
)

type DataHandlerFunc func(ctx context.Context, d *Data)

type Data struct {
	Feed               Feed   `json:"feed"`
	OriginalXMLMessage string `json:"original_xml_message,omitempty"`
}

func (d *Data) String() string {
	tmp := d.OriginalXMLMessage
	d.OriginalXMLMessage = ""
	str, _ := json.Marshal(d)
	d.OriginalXMLMessage = tmp

	return string(str)
}

type Feed struct {
	YT           string       `json:"-yt,omitempty"`
	XMLNS        string       `json:"-xmlns"`
	DeletedEntry DeletedEntry `json:"deleted-entry,omitempty"`
	At           string       `json:"-at,omitempty"`
	Title        string       `json:"title,omitempty"`
	Updated      string       `json:"updated,omitempty"` // 2020-07-29T10:12:08.794405158+00:00
	Link         []Link       `json:"link,omitempty"`
	Entry        Entry        `json:"entry,omitempty"`
}

type DeletedEntry struct {
	Ref  string `json:"-ref,omitempty"`
	When string `json:"-when,omitempty"` // "2020-07-29T16:46:32+00:00"
	Link Link   `json:"link,omitempty"`
	By   Author `json:"by,omitempty"`
}

type Link struct {
	Rel  string `json:"-rel,omitempty"`
	Href string `json:"-href"`
}

type Entry struct {
	Title     string `json:"title"`
	Link      Link   `json:"link"`
	Author    Author `json:"author"`
	Published string `json:"published"` // 2020-07-29T10:12:08.794405158+00:00
	Updated   string `json:"updated"`   // 2020-07-29T10:12:08.794405158+00:00
	ID        string `json:"id"`
	VideoID   string `json:"videoId"`
	ChannelID string `json:"channelId"`
}

type Author struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}
