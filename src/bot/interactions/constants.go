package interactions

import "github.com/bwmarrin/discordgo"

var GUILDS_ONLY []discordgo.ChannelType = []discordgo.ChannelType{
	discordgo.ChannelTypeGuildText,
	discordgo.ChannelTypeGuildVoice,
	discordgo.ChannelTypeGuildCategory,
	discordgo.ChannelTypeGuildNews,
	discordgo.ChannelTypeGuildStore,
	discordgo.ChannelTypeGuildNewsThread,
	discordgo.ChannelTypeGuildPublicThread,
	discordgo.ChannelTypeGuildPrivateThread,
	discordgo.ChannelTypeGuildStageVoice,
	discordgo.ChannelTypeGuildDirectory,
	discordgo.ChannelTypeGuildForum,
	discordgo.ChannelTypeGuildMedia,
	discordgo.ChannelTypeGuildStageVoice,
	discordgo.ChannelTypeGuildForum,
}

// these are the only auto-archive durations allowed by discord
// https://discord.com/developers/docs/resources/channel#thread-metadata-object-thread-auto-archive-duration
// idk why discordgo doesn't have these defined
const (
    AUTO_ARCHIVE_1_HOUR int = 60
    AUTO_ARCHIVE_1_DAY int = 1440
    AUTO_ARCHIVE_3_DAYS int = 4320
	AUTO_ARCHIVE_1_WEEK int = 10080
)
