package segment

import (
	"fmt"

	"github.com/segmentio/analytics-go/v3"
)

const ORG_TRAIT_NAME = "orgName"

var userId string

func NewUser(email string, org string) error {
	var err error

	SetUser(email)

	user := analytics.Identify{
		UserId: email,
		Traits: analytics.NewTraits().SetEmail(email).Set(ORG_TRAIT_NAME, org),
	}

	tenantUniqueId := fmt.Sprintf("%s@%s", org, org)
	orgGroup := analytics.Group{
		GroupId: tenantUniqueId,
		UserId:  user.UserId,
		Traits:  analytics.NewTraits().SetName(tenantUniqueId),
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
