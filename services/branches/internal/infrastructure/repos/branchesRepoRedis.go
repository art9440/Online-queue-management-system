package repos

import "github.com/redis/go-redis/v9"

type RegistrationRepoRedis struct {
	client *redis.Client
}

func NewRegistrationRepoRedis(client *redis.Client) *RegistrationRepoRedis {
	return &RegistrationRepoRedis{
		client: client,
	}
}
