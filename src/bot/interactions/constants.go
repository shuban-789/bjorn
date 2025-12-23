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
