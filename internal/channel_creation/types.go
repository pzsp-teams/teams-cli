package channelcreation

type CreateChannelBody struct {
	TeamRef string
	ChannelRef string
	MemberRefs []string
	OwnerRefs []string
}