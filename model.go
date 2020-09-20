package ytfeed

import "time"

type YTFeedModel struct {
	ID          string    `db:"id"`
	PublishedAt time.Time `db:"published_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	VideoID     string    `db:"video_id"`
	ChannelID   string    `db:"channel_id"`
	Title       string    `db:"title"`
	URL         string    `db:"url"`
}
