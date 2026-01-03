package testutil

//go:generate mockgen -destination=mock_channels_service.go -package=testutil -mock_names=Service=MockChannelsService github.com/pzsp-teams/lib/channels Service
//go:generate mockgen -destination=mock_chats_service.go -package=testutil -mock_names=Service=MockChatsService github.com/pzsp-teams/lib/chats Service
//go:generate mockgen -destination=mock_sender_adapter.go -package=testutil -source=../messaging/sender_adapter.go senderAdapter
