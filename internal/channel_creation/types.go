package channelcreation

type createChannelBody struct {
	TeamRef    string
	ChannelRef string
	MemberRefs []string
	OwnerRefs  []string
}
