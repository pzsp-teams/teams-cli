package retriever

//go:generate mockgen -destination=mock_teams_service_test.go -package=retriever -mock_names=Service=mockTeamsService github.com/pzsp-teams/lib/teams Service
//go:generate mockgen -destination=mock_channels_service_test.go -package=retriever -mock_names=Service=mockChannelsService github.com/pzsp-teams/lib/channels Service
