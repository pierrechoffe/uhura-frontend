package core

import (
	. "github.com/fiam/gounidecode/unidecode"
	"github.com/jinzhu/gorm"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Channel struct {
	Title         string `sql:"not null;unique"`
	Description   string
	ImageUrl      string
	Copyright     string
	LastBuildDate string
	Url           string `sql:"not null;unique"`
	Id            int
	Uri           string
	Featured      bool
}

type UserChannel struct {
	Id        int
	UserId    int
	ChannelId int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChannelResult struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ImageUrl    string      `json:"image_url"`
	Url         string      `json:"url"`
	Id          int         `json:"id"`
	Uri         string      `json:"uri"`
	ToView      int         `json:"to_view"`
	Subscribed  interface{} `json:"subscribed"`
	Copyright   string      `json:"copyright"`
	Episodes    []int64     `json:"episodes"`
}

func (cr *ChannelResult) GetUri() string {
	re := regexp.MustCompile(`\W`)
	uri := Unidecode(cr.Title)
	uri = re.ReplaceAllString(uri, "")
	uri = strings.ToLower(uri)
	uri = strings.Replace(uri, "podcast", "", -1)
	database.Table("channels").Where(cr.Id).Update("Uri", uri)

	return uri
}

func featuredScope(d *gorm.DB) *gorm.DB {
	return d.Not("featured", "false").Order("random()").Limit(12)
}

func userInfoScope(userId string) func(d *gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		return d.Joins("FULL OUTER JOIN user_channels ON user_channels.channel_id=channels.id AND user_channels.user_id=" + userId).Select("channels.*, CAST(user_channels.user_id AS BOOLEAN) AS subscribed ")
	}
}

func AllChannels(userId int, onlyFeatured bool, channelId int) (channels []ChannelResult, episodes []ItemResult) {
	channelQuery := database.Table("channels").Where("title IS NOT NULL").Where("title <> ''")

	if userId > 0 {
		channelQuery = channelQuery.Scopes(userInfoScope(strconv.Itoa(userId)))
	}

	if onlyFeatured {
		channelQuery = channelQuery.Scopes(featuredScope)
	}

	if channelId > 0 {
		channelQuery = channelQuery.Where("channels.id = ?", channelId)
	}

	channelQuery.Order("title").Find(&channels)

	for i, c := range channels {
		if c.Uri == "" {
			c.Uri = c.GetUri()
			channels[i] = c
		}
		var episodesIds []int64
		database.Table("items").Where("channel_id = ?", c.Id).Find(&episodes).Pluck("id", &episodesIds)
		channels[i].Episodes = episodesIds
	}

	return
}

type UserChannelsEntity struct {
	Id        int `json:"id"`
	ChannelId int `json:"channel"`
}

func Subscriptions(user *User) (subscriptions []UserChannelsEntity, channels []ChannelResult) {
	var channelsIds []int64

	database.Table("user_channels").Where("user_id = ?", user.Id).Find(&subscriptions).Pluck("channel_id", &channelsIds)
	database.Table("channels").Where("id IN (?)", channelsIds).Find(&channels)

	for i, channel := range channels {
		var items []struct {
			Id int
		}
		var userItems []interface{}
		var watched []int64

		database.Table("items").Select("DISTINCT items.id").Where("channel_id = ?", channel.Id).Joins("FULL OUTER JOIN user_items ON user_items.item_id=items.id AND user_items.user_id="+strconv.Itoa(user.Id)).Find(&items).Pluck("user_items.id", &userItems)

		for _, j := range userItems {
			id, ok := j.(int64)
			if ok {
				watched = append(watched, id)
			}
		}
		toView := len(items) - len(watched)
		channel.ToView = toView
		channels[i] = channel
	}
	return
}

func SubscribeChannel(userId int, channelId string) (channel ChannelResult) {
	var userChannel UserChannel

	channelIdInt, _ := strconv.Atoi(channelId)

	database.Table("user_channels").Where(UserChannel{ChannelId: channelIdInt, UserId: userId}).FirstOrCreate(&userChannel)
	channels, _ := AllChannels(userId, false, channelIdInt)
	channel = channels[0]
	return
}
