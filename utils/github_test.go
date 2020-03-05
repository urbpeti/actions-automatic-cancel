package utils

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestVerifyGithubWebhookRequest(t *testing.T) {
	t.Run("Missing signature", func(t *testing.T) {
		err := VerifyGithubWebhookRequest(events.APIGatewayProxyRequest{}, "secret")

		if err == nil {
			t.Errorf("Missing error")
		}
		if err.Error() != "Missing signature" {
			t.Errorf("Bad error %s", err.Error())
		}
	})

	t.Run("Bad signature format", func(t *testing.T) {
		err := VerifyGithubWebhookRequest(events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"X-Hub-Signature": "sha1",
			}}, "secret")

		if err == nil {
			t.Errorf("Missing error")
		}
		if err.Error() != "Bad signature format" {
			t.Errorf("Bad error %s", err.Error())
		}
	})

	t.Run("Signature decode err", func(t *testing.T) {
		err := VerifyGithubWebhookRequest(events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"X-Hub-Signature": "sha1=badsign",
			}}, "secret")

		if err == nil {
			t.Errorf("Missing error")
		}
		if err.Error() != "encoding/hex: invalid byte: U+0073 's'" {
			t.Errorf("Bad error %s", err.Error())
		}
	})

	t.Run("Signature missmatch", func(t *testing.T) {
		err := VerifyGithubWebhookRequest(events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"X-Hub-Signature": "sha1=d37e24f84c53a5c2a510694205749219447d494a",
			}}, "secret")

		if err == nil {
			t.Errorf("Missing error")
		}
		if err.Error() != "Signature missmatch" {
			t.Errorf("Bad error %s", err.Error())
		}
	})

	t.Run("Valid signature", func(t *testing.T) {
		err := VerifyGithubWebhookRequest(events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"X-Hub-Signature": "sha1=2486c8590c396f876a46fb541e57fb3f9f276052",
			},
			Body: "dummy",
		}, "secret")

		if err != nil {
			t.Errorf("Should not return error %s", err.Error())
		}
	})
}
