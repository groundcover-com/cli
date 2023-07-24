package segment

import (
	"crypto/sha256"
	"fmt"

	"github.com/segmentio/analytics-go/v3"
)

const ORG_TRAIT_NAME = "orgName"

var userId string

func NewUser(email string, org string) error {
	var err error

	userId := fmt.Sprintf("%x", sha256.Sum256([]byte(email)))

	SetUser(userId)

	user := analytics.Identify{
		UserId: userId,
		Traits: analytics.NewTraits().SetEmail(email).Set(ORG_TRAIT_NAME, org),
	}

	tenantUniqueId := fmt.Sprintf("%s@%s", org, org)
	orgGroup := analytics.Group{
		GroupId: tenantUniqueId,
		UserId:  userId,
		Traits:  analytics.NewTraits().SetEmail(email).SetName(tenantUniqueId),
	}

	if err = client.Enqueue(user); err != nil {
		return err
	}

	if err = client.Enqueue(orgGroup); err != nil {
		return err
	}

	return nil
}

func SetUser(email string) {
	userId = email
}
